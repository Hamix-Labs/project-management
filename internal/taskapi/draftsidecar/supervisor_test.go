package draftsidecar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	draftassistmetrics "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/metrics"
)

// fakeChild is a childProcess whose stdout is a caller-controlled pipe.
// Wait blocks until Kill is called or exit is signalled.
type fakeChild struct {
	stdout io.ReadCloser
	closed chan struct{}
	once   sync.Once
	exit   error
	killed atomic.Bool
}

func newFakeChild(stdout io.ReadCloser) *fakeChild {
	return &fakeChild{stdout: stdout, closed: make(chan struct{})}
}

func (c *fakeChild) Stdout() io.ReadCloser { return c.stdout }
func (c *fakeChild) Wait() error {
	<-c.closed
	return c.exit
}
func (c *fakeChild) Kill() error {
	c.once.Do(func() {
		c.killed.Store(true)
		close(c.closed)
		if c.stdout != nil {
			_ = c.stdout.Close()
		}
	})
	return nil
}
func (c *fakeChild) exitWith(err error) {
	c.once.Do(func() {
		c.exit = err
		close(c.closed)
		if c.stdout != nil {
			_ = c.stdout.Close()
		}
	})
}

// spawnScript builds a spawn func that returns children off a channel in
// FIFO order, so tests can inject a specific respawn sequence.
type spawnScript struct {
	children []*fakeChild
	envSeen  [][]string
	argsSeen [][]string
	mu       sync.Mutex
	idx      int
}

func newSpawnScript(children ...*fakeChild) *spawnScript {
	return &spawnScript{children: children}
}

func (s *spawnScript) spawn(_ context.Context, _ string, args []string, env []string, _ io.Writer) (childProcess, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.idx >= len(s.children) {
		return nil, fmt.Errorf("test: no more scripted children (index %d)", s.idx)
	}
	c := s.children[s.idx]
	s.idx++
	s.argsSeen = append(s.argsSeen, args)
	s.envSeen = append(s.envSeen, env)
	return c, nil
}

// pipePair returns a piped stdout the test can write to. Newlines matter
// because bufio.Scanner reads line-by-line.
func pipePair() (io.ReadCloser, io.WriteCloser) {
	r, w := io.Pipe()
	return r, w
}

func writeListenLine(w io.WriteCloser, port int) {
	_, _ = w.Write([]byte("listening on " + fmtInt(port) + "\n"))
}

func fmtInt(n int) string { return fmt.Sprintf("%d", n) }

// makeReadyzServer returns an httptest.Server whose /readyz replies with
// the supplied JSON string. Non-/readyz paths return 404.
func makeReadyzServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/readyz" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestReadListenPort_ParsesFirstMatchingLine(t *testing.T) {
	r, w := pipePair()
	go func() {
		_, _ = w.Write([]byte("boot: starting\n"))
		_, _ = w.Write([]byte("listening on 51234\n"))
		_, _ = w.Write([]byte("noise\n"))
		_ = w.Close()
	}()
	port, err := readListenPort(r)
	if err != nil {
		t.Fatalf("readListenPort: %v", err)
	}
	if port != 51234 {
		t.Fatalf("port=%d want 51234", port)
	}
}

func TestReadListenPort_FailsWhenClosedBeforeLine(t *testing.T) {
	r, w := pipePair()
	_ = w.Close()
	if _, err := readListenPort(r); err == nil {
		t.Fatal("expected error on closed stdout")
	}
}

func TestSupervisor_StartMissingKey_Errors(t *testing.T) {
	sup := NewSupervisor(Options{
		BinaryPath: "unused",
		Logger:     slog.Default(),
	})
	if err := sup.Start(context.Background()); err == nil {
		t.Fatal("Start should have failed without API key")
	}
	ready, name, reason := sup.Ready()
	if ready || name != "sdk" || reason != draftassistmetrics.ReasonMissingKey {
		t.Fatalf("ready=%v name=%q reason=%q", ready, name, reason)
	}
}

// TestSupervisor_ParsesPortAndReportsReady wires a fake child that prints
// a listen line matching an httptest /readyz endpoint. The probe should
// flip to ready=true within one interval.
func TestSupervisor_ParsesPortAndReportsReady(t *testing.T) {
	srv := makeReadyzServer(t, `{"ready":true}`)
	port := extractPort(t, srv.URL)

	stdout, stdin := pipePair()
	child := newFakeChild(stdout)
	script := newSpawnScript(child)

	sup := newSupervisorForTest(t, "hamix-draft-agent", script, 20*time.Millisecond)
	go writeListenLine(stdin, port)

	if err := sup.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = sup.Close() })

	if got := sup.Port(); got != port {
		t.Fatalf("port=%d want %d", got, port)
	}
	// Poll for ready to flip; probe interval is 20ms.
	waitFor(t, 500*time.Millisecond, func() bool {
		ready, _, _ := sup.Ready()
		return ready
	})
}

// TestSupervisor_RestartsWithBackoff kills the initial child and asserts
// that the supervisor respawns a fresh one within the first backoff
// window (which we compress to milliseconds via the test hook).
func TestSupervisor_RestartsWithBackoff(t *testing.T) {
	srv := makeReadyzServer(t, `{"ready":true}`)
	port := extractPort(t, srv.URL)

	child1Stdout, child1Stdin := pipePair()
	child2Stdout, child2Stdin := pipePair()
	child1 := newFakeChild(child1Stdout)
	child2 := newFakeChild(child2Stdout)
	script := newSpawnScript(child1, child2)

	sup := newSupervisorForTest(t, "hamix-draft-agent", script, 10*time.Millisecond)
	// Compress backoffs so the test finishes fast.
	sup.opts.backoffs = []time.Duration{5 * time.Millisecond, 10 * time.Millisecond}

	go writeListenLine(child1Stdin, port)
	if err := sup.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = sup.Close() })

	waitFor(t, 500*time.Millisecond, func() bool {
		ready, _, _ := sup.Ready()
		return ready
	})

	// Simulate crash: the first child exits, the supervisor should
	// respawn (using child2 next).
	go writeListenLine(child2Stdin, port)
	child1.exitWith(errors.New("boom"))

	waitFor(t, 1*time.Second, func() bool {
		sup.mu.RLock()
		restarts := sup.restarts
		sup.mu.RUnlock()
		return restarts >= 1
	})
	if !child1.killed.Load() && child1.exit == nil {
		t.Fatalf("child1 not stopped")
	}
}

// TestSupervisor_ReadyReportsSidecarDown_WhenProbeFails wires a fake
// child but no /readyz — the probe should keep reporting sidecar_down.
func TestSupervisor_ReadyReportsSidecarDown_WhenProbeFails(t *testing.T) {
	stdout, stdin := pipePair()
	child := newFakeChild(stdout)
	script := newSpawnScript(child)

	// Choose a port unlikely to have anything listening.
	sup := newSupervisorForTest(t, "hamix-draft-agent", script, 10*time.Millisecond)
	go writeListenLine(stdin, 1)

	if err := sup.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = sup.Close() })

	// Wait past one probe interval so probeOnce has executed.
	time.Sleep(50 * time.Millisecond)
	ready, name, reason := sup.Ready()
	if ready {
		t.Fatalf("ready should be false")
	}
	if name != "sdk" {
		t.Fatalf("name=%q want sdk", name)
	}
	if reason != draftassistmetrics.ReasonSidecarDown {
		t.Fatalf("reason=%q want %q", reason, draftassistmetrics.ReasonSidecarDown)
	}
}

func TestSupervisor_ChildEnvContainsAPIKey_AndStripsExisting(t *testing.T) {
	stdout, _ := pipePair()
	child := newFakeChild(stdout)
	script := newSpawnScript(child)

	sup := NewSupervisor(Options{
		BinaryPath: "hamix-draft-agent",
		APIKey:     "sekrit",
		Env: []string{
			"CURSOR_API_KEY=stale-should-be-stripped",
			"PATH=/x",
			"HOME=/home/user",
		},
		Logger: slog.Default(),
	})
	sup.opts.spawn = script.spawn
	sup.opts.backoffs = []time.Duration{time.Millisecond}
	sup.opts.ProbeInterval = time.Second
	sup.opts.StartupTimeout = 200 * time.Millisecond

	// Start will time out waiting for the listen line — that's fine, we
	// only care about the env captured by spawn.
	_ = sup.Start(context.Background())

	if len(script.envSeen) == 0 {
		t.Fatal("spawn env not captured")
	}
	env := script.envSeen[0]
	var count int
	for _, kv := range env {
		if strings.HasPrefix(kv, "CURSOR_API_KEY=") {
			count++
			if kv != "CURSOR_API_KEY=sekrit" {
				t.Fatalf("api key kv=%q", kv)
			}
		}
		if strings.Contains(kv, "stale-should-be-stripped") {
			t.Fatalf("stale api key leaked through: %q", kv)
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one CURSOR_API_KEY entry, got %d", count)
	}
}

// newSupervisorForTest returns a supervisor with all defaults except the
// spawn hook and probe interval. Test uses os.Setenv-less API key.
func newSupervisorForTest(t *testing.T, binary string, script *spawnScript, probeInterval time.Duration) *Supervisor {
	t.Helper()
	sup := NewSupervisor(Options{
		BinaryPath: binary,
		APIKey:     "test-key",
		Env:        []string{"PATH=/x"},
		Logger:     slog.Default(),
	})
	sup.opts.spawn = script.spawn
	sup.opts.ProbeInterval = probeInterval
	sup.opts.ReadyTimeout = 100 * time.Millisecond
	sup.opts.StartupTimeout = 1 * time.Second
	return sup
}

func extractPort(t *testing.T, url string) int {
	t.Helper()
	// url looks like http://127.0.0.1:<port>
	idx := strings.LastIndex(url, ":")
	if idx < 0 {
		t.Fatalf("bad url %q", url)
	}
	var port int
	if _, err := fmt.Sscanf(url[idx+1:], "%d", &port); err != nil {
		t.Fatalf("parse port %q: %v", url, err)
	}
	return port
}

func waitFor(t *testing.T, d time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("waitFor: condition never true")
}
