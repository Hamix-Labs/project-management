package draftsidecar

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"

	draftassistmetrics "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/metrics"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// BinaryName is the executable name the supervisor searches for on PATH.
const BinaryName = "hamix-draft-agent"

// APIKeyEnv names the environment variable that carries the Cursor SDK
// authentication token. The supervisor never logs the value; it only
// propagates it to the child via os.Environ() (see spawnCmd).
const APIKeyEnv = "CURSOR_API_KEY"

const (
	defaultProbeInterval   = 2 * time.Second
	defaultReadyTimeout    = 800 * time.Millisecond
	defaultShutdownTimeout = 3 * time.Second
	defaultStartupTimeout  = 10 * time.Second
)

var (
	backoffSchedule = []time.Duration{1 * time.Second, 2 * time.Second, 5 * time.Second, 15 * time.Second}

	listenLine = regexp.MustCompile(`^listening on (\d+)\s*$`)
)

// Options tunes the supervisor. All fields are optional; the zero value
// falls back to production defaults suitable for taskapi boot.
type Options struct {
	// BinaryPath overrides exec.LookPath(BinaryName) when set.
	BinaryPath string
	// APIKey is the value to forward as CURSOR_API_KEY. When empty and
	// os.Getenv returns empty, the supervisor refuses to start.
	APIKey string
	// Env is the base environment for the child. Defaults to os.Environ()
	// filtered to remove any duplicate CURSOR_API_KEY that would otherwise
	// leak through unchanged.
	Env []string
	// Stderr receives the child's stderr. Nil discards.
	Stderr io.Writer
	// Logger is the slog handle used for supervisor lifecycle messages;
	// slog.Default is used when nil.
	Logger *slog.Logger
	// ProbeInterval is the delay between /readyz checks. Defaults to 2s.
	ProbeInterval time.Duration
	// ReadyTimeout is the per-probe HTTP timeout. Defaults to 800ms.
	ReadyTimeout time.Duration
	// StartupTimeout bounds how long Start waits for the port line.
	StartupTimeout time.Duration
	// spawn is a test seam: replaces exec.Command construction. Set only
	// in tests via TestingHooks.
	spawn spawnFunc
	// backoffs overrides backoffSchedule for tests.
	backoffs []time.Duration
	// httpClient is the *http.Client used to probe /readyz. Defaults to a
	// short-timeout loopback client. Test seam.
	httpClient *http.Client
	// nowFn is a clock override for tests. Defaults to time.Now.
	nowFn func() time.Time
}

// spawnFunc returns a started *exec.Cmd-like handle. Tests substitute this
// to avoid touching real processes.
type spawnFunc func(ctx context.Context, bin string, args []string, env []string, stderr io.Writer) (childProcess, error)

// childProcess is the narrow surface the supervisor needs from the child.
// *exec.Cmd is adapted via execChild below; tests provide a fake.
type childProcess interface {
	Stdout() io.ReadCloser
	Wait() error
	Kill() error
}

// Supervisor owns the sidecar process lifetime and answers ReadyProbe.
//
// Zero-value Supervisor is not usable — call NewSupervisor. The type is
// safe for concurrent Ready() / Port() calls; Start / Close are single
// shot.
type Supervisor struct {
	opts       Options
	log        *slog.Logger
	httpClient *http.Client
	nowFn      func() time.Time

	// runCtx / runCancel scope the supervisor loop; Close cancels them.
	runCtx    context.Context
	runCancel context.CancelFunc
	loopDone  chan struct{}

	// State guarded by mu. Ready() reads a snapshot.
	mu       sync.RWMutex
	port     int
	up       bool
	reason   string
	restarts int
	closed   bool

	// hasKey is captured at construction; Ready() reports missing_key
	// without touching the process.
	hasKey bool
}

// NewSupervisor builds a supervisor for the sidecar binary. It does not
// spawn the child — call Start(ctx) once boot ordering is right.
//
// The API key is captured from opts.APIKey and, when that is empty, from
// os.Getenv(APIKeyEnv). The captured value is only ever placed on the
// child's environment; the supervisor never logs it.
//
//funclogmeasure:skip category=tool-required-noop reason="Constructor; lifecycle traces live on Start/Close and the run loop."
func NewSupervisor(opts Options) *Supervisor {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.ProbeInterval <= 0 {
		opts.ProbeInterval = defaultProbeInterval
	}
	if opts.ReadyTimeout <= 0 {
		opts.ReadyTimeout = defaultReadyTimeout
	}
	if opts.StartupTimeout <= 0 {
		opts.StartupTimeout = defaultStartupTimeout
	}
	if opts.spawn == nil {
		opts.spawn = defaultSpawn
	}
	if len(opts.backoffs) == 0 {
		opts.backoffs = backoffSchedule
	}
	if opts.nowFn == nil {
		opts.nowFn = time.Now
	}
	key := opts.APIKey
	if key == "" {
		key = os.Getenv(APIKeyEnv)
	}
	s := &Supervisor{
		opts:       opts,
		log:        opts.Logger,
		httpClient: opts.httpClient,
		nowFn:      opts.nowFn,
		hasKey:     key != "",
		reason:     draftassistmetrics.ReasonSidecarDown,
	}
	if !s.hasKey {
		s.reason = draftassistmetrics.ReasonMissingKey
	}
	if s.httpClient == nil {
		s.httpClient = &http.Client{Timeout: opts.ReadyTimeout}
	}
	// Bake the resolved key back onto opts so spawnCmd sees a single source.
	s.opts.APIKey = key
	return s
}

// Start spawns the sidecar and blocks until the child prints the
// "listening on <port>" line (or StartupTimeout elapses). Once the port
// is known, Start launches the supervise + probe loops and returns.
//
//funclogmeasure:skip category=tool-required-noop reason="Boot-time supervisor start; startup log emitted below."
func (s *Supervisor) Start(ctx context.Context) error {
	if !s.hasKey {
		return fmt.Errorf("draftsidecar: %s is not set", APIKeyEnv)
	}
	bin := s.opts.BinaryPath
	if bin == "" {
		found, err := exec.LookPath(BinaryName)
		if err != nil {
			return fmt.Errorf("draftsidecar: locate %s: %w", BinaryName, err)
		}
		bin = found
	}
	s.opts.BinaryPath = bin

	s.mu.Lock()
	if s.runCtx != nil {
		s.mu.Unlock()
		return errors.New("draftsidecar: supervisor already started")
	}
	s.runCtx, s.runCancel = context.WithCancel(context.Background())
	s.loopDone = make(chan struct{})
	s.mu.Unlock()

	// Spawn synchronously so callers see an early error if the child
	// refuses to boot on first attempt. The child's context is runCtx
	// so it lives past the caller's Start ctx; the startup wait honors
	// the caller's ctx for cancellation.
	child, port, err := s.spawnAndWaitForPortWithCtx(s.runCtx, ctx)
	if err != nil {
		s.runCancel()
		close(s.loopDone)
		return err
	}
	s.setPort(port)
	s.log.Info("draftsidecar started",
		"cmd", calltrace.LogCmd,
		"operation", "draftsidecar.Supervisor.Start",
		"port", port,
	)

	go s.run(child)
	return nil
}

// Close terminates the child (SIGKILL fallback after shutdownTimeout) and
// waits for the supervise loop to exit. Safe to call once.
//
//funclogmeasure:skip category=tool-required-noop reason="Shutdown hook; child exit is logged in the run loop."
func (s *Supervisor) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	cancel := s.runCancel
	done := s.loopDone
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		select {
		case <-done:
		case <-time.After(defaultShutdownTimeout + 2*time.Second):
			return errors.New("draftsidecar: supervise loop did not exit")
		}
	}
	draftassistmetrics.SetSidecarUp(false)
	return nil
}

// Port returns the port the current child is listening on, or 0 before
// Start has parsed the listen line.
//
//funclogmeasure:skip category=hot-path reason="Accessor; no I/O."
func (s *Supervisor) Port() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
}

// Ready implements draftassist/handler.ReadyProbe. runner is always "sdk"
// while the supervisor is wired; reason follows the constants in
// pkgs/draftassist/metrics.
//
//funclogmeasure:skip category=hot-path reason="Ready-probe accessor; the /draft-assist/ready handler already traces."
func (s *Supervisor) Ready() (bool, string, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.hasKey {
		return false, "sdk", draftassistmetrics.ReasonMissingKey
	}
	if s.up {
		return true, "sdk", ""
	}
	return false, "sdk", draftassistmetrics.ReasonSidecarDown
}

// run is the top-level supervise loop. It runs one child at a time,
// probing readiness while alive; on exit it applies exponential backoff
// then respawns.
func (s *Supervisor) run(initial childProcess) {
	defer close(s.loopDone)

	child := initial
	backoffIdx := 0
	for {
		exit := s.superviseOne(child)
		if s.runCtx.Err() != nil {
			s.killIfAlive(child)
			return
		}
		s.log.Warn("draftsidecar child exited",
			"cmd", calltrace.LogCmd,
			"operation", "draftsidecar.Supervisor.run",
			"err", exitErrMsg(exit),
		)
		s.markDown()

		delay := s.opts.backoffs[minInt(backoffIdx, len(s.opts.backoffs)-1)]
		if !sleepCtx(s.runCtx, delay) {
			return
		}
		backoffIdx++

		next, port, err := s.spawnAndWaitForPortWithCtx(s.runCtx, s.runCtx)
		if err != nil {
			s.log.Warn("draftsidecar respawn failed",
				"cmd", calltrace.LogCmd,
				"operation", "draftsidecar.Supervisor.run",
				"err", err,
			)
			continue
		}
		s.setPort(port)
		s.mu.Lock()
		s.restarts++
		s.mu.Unlock()
		draftassistmetrics.IncSidecarRestart()
		s.log.Info("draftsidecar respawned",
			"cmd", calltrace.LogCmd,
			"operation", "draftsidecar.Supervisor.run",
			"port", port,
			"restarts", s.restarts,
		)
		backoffIdx = 0
		child = next
	}
}

// superviseOne probes /readyz until the child exits or the supervisor is
// closed. Returns the child's exit error (if any).
func (s *Supervisor) superviseOne(child childProcess) error {
	waitDone := make(chan error, 1)
	go func() { waitDone <- child.Wait() }()

	ticker := time.NewTicker(s.opts.ProbeInterval)
	defer ticker.Stop()

	s.probeOnce()

	for {
		select {
		case <-s.runCtx.Done():
			_ = child.Kill()
			return <-waitDone
		case err := <-waitDone:
			return err
		case <-ticker.C:
			s.probeOnce()
		}
	}
}

func (s *Supervisor) probeOnce() {
	port := s.Port()
	if port == 0 {
		s.markDown()
		return
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/readyz", port)
	ctx, cancel := context.WithTimeout(s.runCtx, s.opts.ReadyTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.markDown()
		return
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.markDown()
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		s.markDown()
		return
	}
	var body struct {
		Ready  bool   `json:"ready"`
		Reason string `json:"reason,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		s.markDown()
		return
	}
	if !body.Ready {
		s.markReason(mapSidecarReason(body.Reason))
		return
	}
	s.markUp()
}

func (s *Supervisor) setPort(port int) {
	s.mu.Lock()
	s.port = port
	s.mu.Unlock()
}

func (s *Supervisor) markUp() {
	s.mu.Lock()
	changed := !s.up
	s.up = true
	s.reason = ""
	s.mu.Unlock()
	if changed {
		draftassistmetrics.SetSidecarUp(true)
	}
}

func (s *Supervisor) markDown() {
	s.mu.Lock()
	changed := s.up
	s.up = false
	if s.reason == "" {
		s.reason = draftassistmetrics.ReasonSidecarDown
	}
	s.mu.Unlock()
	if changed {
		draftassistmetrics.SetSidecarUp(false)
	}
}

func (s *Supervisor) markReason(reason string) {
	s.mu.Lock()
	s.up = false
	s.reason = reason
	s.mu.Unlock()
	draftassistmetrics.SetSidecarUp(false)
}

// spawnAndWaitForPortWithCtx starts a child process (whose lifetime is
// bound to childCtx) and blocks until the listen line appears on stdout
// or StartupTimeout / waitCtx expire. The returned child owns the stdout
// pipe; the caller must eventually Wait/Kill.
func (s *Supervisor) spawnAndWaitForPortWithCtx(childCtx, waitCtx context.Context) (childProcess, int, error) {
	env := s.childEnv()
	child, err := s.opts.spawn(childCtx, s.opts.BinaryPath, []string{"--port", "0"}, env, s.opts.Stderr)
	if err != nil {
		return nil, 0, fmt.Errorf("draftsidecar: spawn: %w", err)
	}
	portCh := make(chan int, 1)
	errCh := make(chan error, 1)
	go func() {
		port, err := readListenPort(child.Stdout())
		if err != nil {
			errCh <- err
			return
		}
		portCh <- port
	}()

	timer := time.NewTimer(s.opts.StartupTimeout)
	defer timer.Stop()
	select {
	case port := <-portCh:
		return child, port, nil
	case err := <-errCh:
		_ = child.Kill()
		return nil, 0, err
	case <-timer.C:
		_ = child.Kill()
		return nil, 0, fmt.Errorf("draftsidecar: timed out waiting for %q line", "listening on <port>")
	case <-waitCtx.Done():
		_ = child.Kill()
		return nil, 0, waitCtx.Err()
	}
}

func (s *Supervisor) childEnv() []string {
	base := s.opts.Env
	if base == nil {
		base = os.Environ()
	}
	// Strip any existing CURSOR_API_KEY so we control the value passed
	// to the child (Options.APIKey / os.Getenv result) exactly once. We
	// never emit the value to logs or errors.
	filtered := make([]string, 0, len(base)+1)
	prefix := APIKeyEnv + "="
	for _, kv := range base {
		if len(kv) >= len(prefix) && kv[:len(prefix)] == prefix {
			continue
		}
		filtered = append(filtered, kv)
	}
	filtered = append(filtered, prefix+s.opts.APIKey)
	return filtered
}

func (s *Supervisor) killIfAlive(child childProcess) {
	if child == nil {
		return
	}
	_ = child.Kill()
}

// readListenPort reads the child stdout line-by-line, looking for a line
// matching `^listening on (\d+)$`. On success the port is returned and
// the reader is drained in a background goroutine so the pipe never
// blocks. Any unrelated lines are ignored (they may show up as sidecar
// info logs before the listen line).
func readListenPort(r io.ReadCloser) (int, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if m := listenLine.FindStringSubmatch(line); len(m) == 2 {
			port, err := strconv.Atoi(m[1])
			if err != nil {
				return 0, fmt.Errorf("draftsidecar: parse port %q: %w", m[1], err)
			}
			// Drain any remaining stdout so the pipe never blocks the
			// child. Errors here are ignored; the child owns the fd.
			go func() { _, _ = io.Copy(io.Discard, r) }()
			return port, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("draftsidecar: read stdout: %w", err)
	}
	return 0, errors.New("draftsidecar: stdout closed before listen line")
}

// mapSidecarReason coerces a /readyz reason string into one of the
// canonical draftassistmetrics.Reason* constants.
func mapSidecarReason(raw string) string {
	switch raw {
	case draftassistmetrics.ReasonMissingKey:
		return draftassistmetrics.ReasonMissingKey
	case draftassistmetrics.ReasonSidecarDown:
		return draftassistmetrics.ReasonSidecarDown
	case "":
		return draftassistmetrics.ReasonSidecarDown
	default:
		return draftassistmetrics.ReasonSidecarDown
	}
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func exitErrMsg(err error) string {
	if err == nil {
		return "exit ok"
	}
	return err.Error()
}

// --- default exec-based spawn ------------------------------------------------

type execChild struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

func (c *execChild) Stdout() io.ReadCloser { return c.stdout }
func (c *execChild) Wait() error           { return c.cmd.Wait() }
func (c *execChild) Kill() error {
	if c.cmd.Process == nil {
		return nil
	}
	return c.cmd.Process.Kill()
}

func defaultSpawn(ctx context.Context, bin string, args []string, env []string, stderr io.Writer) (childProcess, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = env
	if stderr != nil {
		cmd.Stderr = stderr
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		return nil, err
	}
	return &execChild{cmd: cmd, stdout: stdout}, nil
}
