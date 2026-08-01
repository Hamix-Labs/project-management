package handlertest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/gittest"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"

	// Register production runners so POST /tasks can resolve settings.Runner.
	_ "github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/registry/all"
)

// GitBinding holds the git repository/project/worktree IDs seeded for a test server.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git binding; not part of production trace paths."
type GitBinding struct {
	RepositoryID string
	ProjectID    string
	WorktreeID   string
}

// DirectHandlerTestURL is the base URL used when exercising a handler directly
// via httptest.NewRecorder (no real HTTP server). Tests must register a git
// binding under this URL via NewDirectBoundHandler so WithCreateChecklistForURL
// injects repository_id and project_id.
const DirectHandlerTestURL = "http://handler-direct.test"

// TestCriterionText is the default non-empty done criterion injected by
// WithCreateChecklistForURL into POST /tasks bodies.
const TestCriterionText = "Test criterion"

// JSONErrorBody is the standard API error envelope used in middleware and
// handler contract tests.
type JSONErrorBody struct {
	Error string `json:"error"`
}

var (
	htSharedBareDir  string
	htSharedGitReady bool
	htSharedGitOnce  sync.Once
	htGitBindings    sync.Map // baseURL → GitBinding
)

// htSharedBareRemote returns one bare origin shared across tests in this
// package. Each test clones into its own workdir so allocate/worktree-add
// does not race on a shared checkout.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only shared git fixture; not part of production trace paths."
func htSharedBareRemote(t *testing.T) string {
	t.Helper()
	htSharedGitOnce.Do(func() {
		if _, err := exec.LookPath("git"); err != nil {
			return
		}
		bare, err := os.MkdirTemp("", "hamix-handlertest-bare-*")
		if err != nil {
			panic("handlertest shared bare: " + err.Error())
		}
		if out, err := exec.Command("git", "init", "--bare", "-b", "main", bare).CombinedOutput(); err != nil {
			panic(fmt.Sprintf("git init --bare: %v %s", err, out))
		}
		seed, err := os.MkdirTemp("", "hamix-handlertest-seed-*")
		if err != nil {
			panic("handlertest seed dir: " + err.Error())
		}
		defer os.RemoveAll(seed)
		if out, err := exec.Command("git", "clone", bare, seed).CombinedOutput(); err != nil {
			panic(fmt.Sprintf("git clone bare: %v %s", err, out))
		}
		for _, args := range [][]string{
			{"config", "user.email", "test@test.local"},
			{"config", "user.name", "Test"},
			{"commit", "--allow-empty", "-m", "init"},
			{"push", "-u", "origin", "main"},
		} {
			if out, err := exec.Command("git", append([]string{"-C", seed}, args...)...).CombinedOutput(); err != nil {
				panic(fmt.Sprintf("git %v: %v %s", args, err, out))
			}
		}
		htSharedBareDir = bare
		htSharedGitReady = true
	})
	if !htSharedGitReady {
		t.Skip("git not on PATH")
	}
	return htSharedBareDir
}

// RegisterGitBinding records a GitBinding under baseURL and removes it when t
// ends. WithCreateChecklistForURL uses this registry.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git binding registry; not part of production trace paths."
func RegisterGitBinding(t *testing.T, baseURL string, binding GitBinding) {
	t.Helper()
	htGitBindings.Store(baseURL, binding)
	t.Cleanup(func() { htGitBindings.Delete(baseURL) })
}

// GitBindingForURL returns the GitBinding registered under baseURL and whether
// one was found.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git binding lookup; not part of production trace paths."
func GitBindingForURL(baseURL string) (GitBinding, bool) {
	v, ok := htGitBindings.Load(baseURL)
	if !ok {
		return GitBinding{}, false
	}
	return v.(GitBinding), true
}

//funclogmeasure:skip category=tool-required-noop reason="Test-only git seed helper; not part of production trace paths."
func htSeedGitRepo(t *testing.T, st *composition.API) GitBinding {
	t.Helper()
	bare := htSharedBareRemote(t)
	parent := t.TempDir()
	dir := filepath.Join(parent, "main")
	if out, err := exec.Command("git", "clone", bare, dir).CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v %s", err, out)
	}
	for _, args := range [][]string{
		{"config", "user.email", "test@test.local"},
		{"config", "user.name", "Test"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	wtID, _ := gittest.SeedWorktree(t, st, dir)
	return htGitBindingFromWorktree(t, st, wtID)
}

//funclogmeasure:skip category=tool-required-noop reason="Test-only git binding mapper; not part of production trace paths."
func htGitBindingFromWorktree(t *testing.T, st *composition.API, worktreeID string) GitBinding {
	t.Helper()
	ctx := context.Background()
	wt, err := st.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil {
		t.Fatalf("GetGitWorktreeByID: %v", err)
	}
	defaultProj, err := st.EnsureGlobalDefaultProject(ctx)
	if err != nil {
		t.Fatalf("EnsureGlobalDefaultProject: %v", err)
	}
	return GitBinding{
		RepositoryID: wt.RepositoryID,
		ProjectID:    defaultProj.ID,
		WorktreeID:   worktreeID,
	}
}

// BoundTaskHandler creates handler.NewHandler with an SSEHub and
// NewSettingsRepoProvider so POST /tasks can resolve git binding fields.
// opts are applied after the default RepoProvider option.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only handler wiring; not part of production trace paths."
func BoundTaskHandler(st *composition.API, opts ...handler.HandlerOption) http.Handler {
	base := []handler.HandlerOption{handler.WithRepoProvider(handler.NewSettingsRepoProvider(st))}
	return handler.NewHandler(st, realtime.NewSSEHub(), nil, append(base, opts...)...)
}

// StartWorktreeProvisioner wires the async allocate loop used in production
// after POST /tasks (ADR-0083). Call on create-capable test stores so ready
// tasks become queue-eligible once the managed worktree binds.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only provisioner wiring; not part of production trace paths."
func StartWorktreeProvisioner(t *testing.T, st *composition.API) {
	t.Helper()
	hub := realtime.NewSSEHub()
	prov := composition.NewWorktreeProvisioner(st, hub)
	st.SetWorktreeProvisioner(prov)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		prov.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		prov.Stop()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	})
}

// NewBoundServer returns an httptest.Server backed by BoundTaskHandler with a
// seeded shared git repo registered under the server URL. build receives the
// composition.API so callers can wrap the handler (e.g. with middleware).
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP server; not part of production trace paths."
func NewBoundServer(t *testing.T, build func(st *composition.API) http.Handler) *httptest.Server {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	st := composition.NewAPI(db)
	binding := htSeedGitRepo(t, st)
	StartWorktreeProvisioner(t, st)
	srv := httptest.NewServer(build(st))
	RegisterGitBinding(t, srv.URL, binding)
	t.Cleanup(srv.Close)
	return srv
}

// NewDirectBoundHandler returns a handler backed by BoundTaskHandler with a
// seeded git repo registered under DirectHandlerTestURL. Use this when tests
// exercise the handler via httptest.NewRecorder rather than http.DefaultClient.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only handler wiring; not part of production trace paths."
func NewDirectBoundHandler(t *testing.T, build func(st *composition.API) http.Handler) http.Handler {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	st := composition.NewAPI(db)
	binding := htSeedGitRepo(t, st)
	StartWorktreeProvisioner(t, st)
	RegisterGitBinding(t, DirectHandlerTestURL, binding)
	return build(st)
}

// WithCreateChecklistForURL injects required checklist_items and, when a git
// binding is registered under baseURL, also repository_id and project_id into a
// POST /tasks JSON object. jsonBody must be a single JSON object literal.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only JSON body builder; not part of production trace paths."
func WithCreateChecklistForURL(baseURL, jsonBody string) string {
	jsonBody = strings.TrimSpace(jsonBody)
	if strings.Contains(jsonBody, "checklist_items") {
		return htWithCreateGitBinding(baseURL, jsonBody)
	}
	if !strings.HasSuffix(jsonBody, "}") {
		return htWithCreateGitBinding(baseURL, jsonBody)
	}
	out := jsonBody[:len(jsonBody)-1] + `,"checklist_items":[{"text":"` + TestCriterionText + `"}]}`
	return htWithCreateGitBinding(baseURL, out)
}

//funclogmeasure:skip category=tool-required-noop reason="Test-only JSON body mutator; not part of production trace paths."
func htWithCreateGitBinding(baseURL, jsonBody string) string {
	jsonBody = strings.TrimSpace(jsonBody)
	binding, ok := GitBindingForURL(baseURL)
	if !ok {
		return jsonBody
	}
	if !strings.HasSuffix(jsonBody, "}") {
		return jsonBody
	}
	var extra []string
	if !strings.Contains(jsonBody, "repository_id") {
		extra = append(extra, `"repository_id":"`+binding.RepositoryID+`"`)
	}
	if !strings.Contains(jsonBody, "project_id") {
		extra = append(extra, `"project_id":"`+binding.ProjectID+`"`)
	}
	// worktree_id is allocated server-side; do not inject a client worktree.
	if len(extra) == 0 {
		return jsonBody
	}
	return jsonBody[:len(jsonBody)-1] + `,` + strings.Join(extra, ",") + `}`
}

// DrainSSE collects up to want events from ch within timeout. It returns
// whatever arrived, plus a short grace-window read to surface unexpected
// extras. Callers should assert on the returned slice.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only SSE drain helper; not part of production trace paths."
func DrainSSE(t *testing.T, ch <-chan string, want int, timeout time.Duration) []realtime.Event {
	t.Helper()
	out := make([]realtime.Event, 0, want)
	deadline := time.After(timeout)
	for len(out) < want {
		select {
		case s, ok := <-ch:
			if !ok {
				return out
			}
			var ev realtime.Event
			if err := json.Unmarshal([]byte(s), &ev); err != nil {
				t.Fatalf("decode sse line %q: %v", s, err)
			}
			out = append(out, ev)
		case <-deadline:
			return out
		}
	}
	// Quick grace-window read to surface unexpected extras the test wasn't expecting.
	select {
	case s := <-ch:
		var ev realtime.Event
		if err := json.Unmarshal([]byte(s), &ev); err == nil {
			out = append(out, ev)
		}
	case <-time.After(50 * time.Millisecond):
	}
	return out
}

// SummarizeSSEEvents collapses a realtime.Event slice into a stable sorted
// string set for comparison without relying on publish order. Format:
// "type:id" for task-only events, "type:id/cycle_id" for task_cycle_changed,
// and "type:id/seq" when event_seq is set.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only SSE summary helper; not part of production trace paths."
func SummarizeSSEEvents(events []realtime.Event) []string {
	out := make([]string, 0, len(events))
	for _, ev := range events {
		if ev.EventSeq > 0 {
			out = append(out, fmt.Sprintf("%s:%s/%d", ev.Type, ev.ID, ev.EventSeq))
			continue
		}
		if ev.CycleID != "" {
			out = append(out, fmt.Sprintf("%s:%s/%s", ev.Type, ev.ID, ev.CycleID))
			continue
		}
		out = append(out, fmt.Sprintf("%s:%s", ev.Type, ev.ID))
	}
	sort.Strings(out)
	return out
}
