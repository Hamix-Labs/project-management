package handlertest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

// NewCreateServer returns an httptest.Server with BoundTaskHandler and a seeded
// git binding registered under the server URL (parity with handler's
// newTaskCreateTestServer).
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP server; not part of production trace paths."
func NewCreateServer(t *testing.T) *httptest.Server {
	t.Helper()
	return NewBoundServer(t, func(st *composition.API) http.Handler {
		return BoundTaskHandler(st)
	})
}

// NewCreateServerWithStore is like [NewCreateServer] but also returns the store.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP server; not part of production trace paths."
func NewCreateServerWithStore(t *testing.T) (*httptest.Server, *composition.API) {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	st := composition.NewAPI(db)
	binding := htSeedGitRepo(t, st)
	srv := httptest.NewServer(BoundTaskHandler(st))
	RegisterGitBinding(t, srv.URL, binding)
	t.Cleanup(srv.Close)
	return srv, st
}

// NewSSETriggerServer returns a create-capable server wired to an explicit
// SSEHub so tests can Subscribe and assert publish triggers.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP/SSE wiring; not part of production trace paths."
func NewSSETriggerServer(t *testing.T) (*httptest.Server, *composition.API, *handler.SSEHub) {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	st := composition.NewAPI(db)
	binding := htSeedGitRepo(t, st)
	hub := handler.NewSSEHub()
	srv := httptest.NewServer(handler.NewHandler(st, hub, nil, handler.WithRepoProvider(handler.NewSettingsRepoProvider(st))))
	RegisterGitBinding(t, srv.URL, binding)
	t.Cleanup(srv.Close)
	return srv, st, hub
}

// WithComposeGitBinding injects repository_id, project_id, and worktree_id
// for compose/draft/template payloads when a binding is registered.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only JSON body mutator; not part of production trace paths."
func WithComposeGitBinding(baseURL, jsonBody string) string {
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
	if !strings.Contains(jsonBody, "worktree_id") {
		extra = append(extra, `"worktree_id":"`+binding.WorktreeID+`"`)
	}
	if len(extra) == 0 {
		return jsonBody
	}
	return jsonBody[:len(jsonBody)-1] + `,` + strings.Join(extra, ",") + `}`
}

// WithComposeChecklistForURL injects checklist_items plus compose git binding
// fields into a compose/template payload JSON object.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only JSON body builder; not part of production trace paths."
func WithComposeChecklistForURL(baseURL, jsonBody string) string {
	jsonBody = strings.TrimSpace(jsonBody)
	if strings.Contains(jsonBody, "checklist_items") {
		return WithComposeGitBinding(baseURL, jsonBody)
	}
	if !strings.HasSuffix(jsonBody, "}") {
		return WithComposeGitBinding(baseURL, jsonBody)
	}
	out := jsonBody[:len(jsonBody)-1] + `,"checklist_items":[{"text":"` + TestCriterionText + `"}]}`
	return WithComposeGitBinding(baseURL, out)
}

// MustCreateTask POSTs jsonBody to /tasks (with checklist + git binding) and
// returns the created task id.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only create helper; not part of production trace paths."
func MustCreateTask(t *testing.T, baseURL, jsonBody string) string {
	t.Helper()
	res, err := http.Post(baseURL+"/tasks", "application/json", strings.NewReader(WithCreateChecklistForURL(baseURL, jsonBody)))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create task status %d body=%s", res.StatusCode, raw)
	}
	var task taskcoredomain.Task
	if err := json.Unmarshal(raw, &task); err != nil {
		t.Fatalf("decode created task: %v body=%s", err, raw)
	}
	return task.ID
}

// MustCreateChecklistTask creates a medium-priority task titled title.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only create helper; not part of production trace paths."
func MustCreateChecklistTask(t *testing.T, srv *httptest.Server, title string) string {
	t.Helper()
	return MustCreateTask(t, srv.URL, `{"title":"`+title+`","priority":"medium"}`)
}

// MustEqualEvents fails when got and want event summary slices differ.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only SSE assertion; not part of production trace paths."
func MustEqualEvents(t *testing.T, route string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %d events %v, want %d %v (docs/api.md trigger table)",
			route, len(got), got, len(want), want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s: event[%d]=%q want %q (full got=%v want=%v)", route, i, got[i], want[i], got, want)
		}
	}
}

// MustHaveTaskUpdatedData fails unless events include task_updated for taskID
// with non-nil Data enrichment.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only SSE assertion; not part of production trace paths."
func MustHaveTaskUpdatedData(t *testing.T, route string, events []realtime.Event, taskID string) {
	t.Helper()
	for _, ev := range events {
		if ev.Type == realtime.TaskUpdated && ev.ID == taskID {
			if ev.Data == nil {
				t.Fatalf("%s: task_updated:%s missing data enrichment (ADR-0026)", route, taskID)
			}
			return
		}
	}
	t.Fatalf("%s: no task_updated:%s in events %v", route, taskID, SummarizeSSEEvents(events))
}

// StartContractServer returns NewServer with Close registered on t.Cleanup.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP server; not part of production trace paths."
func StartContractServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := NewServer(t)
	t.Cleanup(srv.Close)
	return srv
}

// AssertBareError asserts status and exact JSON error string.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only contract helper; not part of production trace paths."
func AssertBareError(t *testing.T, res *http.Response, raw []byte, wantStatus int, wantError string) {
	t.Helper()
	if res.StatusCode != wantStatus {
		t.Fatalf("status %d (want %d) body=%s", res.StatusCode, wantStatus, raw)
	}
	var errBody JSONErrorBody
	if err := json.Unmarshal(raw, &errBody); err != nil {
		t.Fatalf("decode: %v body=%s", err, raw)
	}
	if errBody.Error != wantError {
		t.Fatalf("error=%q want %q", errBody.Error, wantError)
	}
}

// EqualStringSlices reports whether a and b are equal element-wise.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only contract helper; not part of production trace paths."
func EqualStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
