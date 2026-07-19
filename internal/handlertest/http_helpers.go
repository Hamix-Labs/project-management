package handlertest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// MaxTaskEventSeqParamBytes is the path-seg max for event seq query values.
const MaxTaskEventSeqParamBytes = 32

// WithCreateGitBinding injects repository_id and project_id when a binding is registered.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only JSON body mutator; not part of production trace paths."
func WithCreateGitBinding(baseURL, jsonBody string) string {
	return htWithCreateGitBinding(baseURL, jsonBody)
}

// WithCreateChecklist injects checklist_items only (no git binding). Prefer
// WithCreateChecklistForURL when a baseURL binding is registered.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only JSON body builder; not part of production trace paths."
func WithCreateChecklist(jsonBody string) string {
	jsonBody = strings.TrimSpace(jsonBody)
	if strings.Contains(jsonBody, "checklist_items") {
		return jsonBody
	}
	if !strings.HasSuffix(jsonBody, "}") {
		return jsonBody
	}
	return jsonBody[:len(jsonBody)-1] + `,"checklist_items":[{"text":"` + TestCriterionText + `"}]}`
}

// MustCreateTaskForCycles creates a default cycles suite task.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only create helper; not part of production trace paths."
func MustCreateTaskForCycles(t *testing.T, baseURL string) string {
	t.Helper()
	return MustCreateTask(t, baseURL, `{"title":"cycles-task","priority":"medium"}`)
}

// MustCreateTaskBody POSTs body to /tasks and fatals on non-201.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only create helper; not part of production trace paths."
func MustCreateTaskBody(t *testing.T, baseURL, body string) {
	t.Helper()
	res, raw := PostCreate(t, baseURL, body)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create task body=%s status %d resp=%s", body, res.StatusCode, raw)
	}
}

// PostCreate POSTs jsonBody with checklist + git binding applied.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func PostCreate(t *testing.T, baseURL, jsonBody string) (*http.Response, []byte) {
	t.Helper()
	res, err := http.Post(baseURL+"/tasks", "application/json", strings.NewReader(WithCreateChecklistForURL(baseURL, jsonBody)))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	return res, raw
}

// PostCreateRaw POSTs jsonBody without injecting checklist_items.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func PostCreateRaw(t *testing.T, baseURL, jsonBody string) (*http.Response, []byte) {
	t.Helper()
	res, err := http.Post(baseURL+"/tasks", "application/json", strings.NewReader(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	return res, raw
}

// MustGetJSON GETs path and fatals unless status is 200.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func MustGetJSON(t *testing.T, baseURL, path string) ([]byte, *http.Response) {
	t.Helper()
	res, err := http.Get(baseURL + path)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d (want 200) body=%s", path, res.StatusCode, raw)
	}
	return raw, res
}

// AssertHTTPError asserts status and that JSON error contains substr.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only contract helper; not part of production trace paths."
func AssertHTTPError(t *testing.T, res *http.Response, raw []byte, wantStatus int, substr string) {
	t.Helper()
	if res.StatusCode != wantStatus {
		t.Fatalf("status %d (want %d) body=%s", res.StatusCode, wantStatus, raw)
	}
	var errBody JSONErrorBody
	if err := json.Unmarshal(raw, &errBody); err != nil {
		t.Fatalf("decode: %v body=%s", err, raw)
	}
	if !strings.Contains(errBody.Error, substr) {
		t.Fatalf("error=%q want substring %q", errBody.Error, substr)
	}
}

// DoCyclesRequest issues a request with X-Actor: agent.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func DoCyclesRequest(t *testing.T, method, url, body string) (*http.Response, []byte) {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Actor", "agent")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(res.Body)
	if cerr := res.Body.Close(); cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil {
		t.Fatal(err)
	}
	return res, raw
}

// PostCycleJSON POSTs a cycle with X-Actor agent and returns the created cycle id.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func PostCycleJSON(t *testing.T, srv *httptest.Server, taskID, body string, wantStatus int) string {
	t.Helper()
	res, raw := DoCyclesRequest(t, http.MethodPost, srv.URL+"/tasks/"+taskID+"/cycles", body)
	if res.StatusCode != wantStatus {
		t.Fatalf("POST /tasks/%s/cycles status=%d want=%d body=%s", taskID, res.StatusCode, wantStatus, raw)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode created cycle: %v body=%s", err, raw)
	}
	if out.ID == "" {
		t.Fatalf("created cycle missing id: body=%s", raw)
	}
	return out.ID
}

// PostTaskJSON POSTs a task body (with checklist+git) and returns the task.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func PostTaskJSON(t *testing.T, srv *httptest.Server, body string, wantStatus int) taskcoredomain.Task {
	t.Helper()
	res, err := http.Post(srv.URL+"/tasks", "application/json", strings.NewReader(WithCreateChecklistForURL(srv.URL, body)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != wantStatus {
		t.Fatalf("POST /tasks status=%d want=%d body=%s", res.StatusCode, wantStatus, b)
	}
	var out taskcoredomain.Task
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("decode created task: %v body=%s", err, b)
	}
	return out
}

// MustDoJSON issues method against url and fatals unless status matches.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func MustDoJSON(t *testing.T, method, url, body, actor string, wantStatus int) {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if actor != "" {
		req.Header.Set("X-Actor", actor)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, url, res.StatusCode, wantStatus, b)
	}
}

// PatchTaskJSON PATCHes /tasks/{id} and fatals unless status matches.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func PatchTaskJSON(t *testing.T, srv *httptest.Server, id, body string, wantStatus int) {
	t.Helper()
	MustDoJSON(t, http.MethodPatch, srv.URL+"/tasks/"+id, body, "", wantStatus)
}

// GetTask GETs /tasks/{id}.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func GetTask(t *testing.T, baseURL, id string) (*http.Response, []byte) {
	t.Helper()
	res, err := http.Get(baseURL + "/tasks/" + id)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return res, raw
}

// DeleteTask DELETEs /tasks/{id}.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func DeleteTask(t *testing.T, baseURL, id string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, baseURL+"/tasks/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return res, raw
}

// PatchTask PATCHes /tasks/{id} with a JSON body.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func PatchTask(t *testing.T, baseURL, id, body string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPatch, baseURL+"/tasks/"+id, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return res, raw
}

// MustGitBinding returns the GitBinding registered under baseURL or fatals.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only git binding helper; not part of production trace paths."
func MustGitBinding(t *testing.T, baseURL string) GitBinding {
	t.Helper()
	binding, ok := GitBindingForURL(baseURL)
	if !ok {
		t.Fatal("missing git binding")
	}
	return binding
}

// AppendApprovalRequestedEvent appends approval_requested and returns its seq.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only event seed helper; not part of production trace paths."
func AppendApprovalRequestedEvent(t *testing.T, st *composition.API, ctx context.Context, taskID string) int64 {
	t.Helper()
	if err := st.AppendTaskEvent(ctx, taskID, taskeventsdomain.EventApprovalRequested, taskcoredomain.ActorAgent, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	seq, err := st.LastEventSeq(ctx, taskID)
	if err != nil {
		t.Fatal(err)
	}
	return seq
}

// FormatEventSeq formats an event seq for path/query use.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only formatting helper; not part of production trace paths."
func FormatEventSeq(seq int64) string {
	return strconv.FormatInt(seq, 10)
}
