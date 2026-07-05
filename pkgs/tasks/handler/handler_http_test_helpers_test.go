package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// postTask POSTs jsonBody to /tasks with checklist and git binding applied.
func postTask(t *testing.T, baseURL, jsonBody string) (*http.Response, []byte) {
	t.Helper()
	res, err := http.Post(baseURL+"/tasks", "application/json", strings.NewReader(withCreateChecklistForURL(baseURL, jsonBody)))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	return res, raw
}

// mustCreateTask POSTs jsonBody to /tasks and returns the created task id.
func mustCreateTask(t *testing.T, baseURL, jsonBody string) string {
	t.Helper()
	res, raw := postTask(t, baseURL, jsonBody)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create task status %d body=%s", res.StatusCode, raw)
	}
	var task domain.Task
	if err := json.Unmarshal(raw, &task); err != nil {
		t.Fatalf("decode created task: %v body=%s", err, raw)
	}
	return task.ID
}

// mustCreateTaskForCycles creates a default-priority task for cycle suite tests.
func mustCreateTaskForCycles(t *testing.T, baseURL string) string {
	t.Helper()
	return mustCreateTask(t, baseURL, `{"title":"cycles-task","priority":"medium"}`)
}

// mustCreateTaskBody POSTs body to /tasks and fatals on non-201 status.
func mustCreateTaskBody(t *testing.T, baseURL, body string) {
	t.Helper()
	res, raw := postTask(t, baseURL, body)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create task body=%s status %d resp=%s", body, res.StatusCode, raw)
	}
}

// postCreate POSTs jsonBody with required checklist_items injected.
func postCreate(t *testing.T, baseURL, jsonBody string) (*http.Response, []byte) {
	t.Helper()
	return postTask(t, baseURL, jsonBody)
}

// postCreateRaw POSTs jsonBody without injecting checklist_items.
func postCreateRaw(t *testing.T, baseURL, jsonBody string) (*http.Response, []byte) {
	t.Helper()
	res, err := http.Post(baseURL+"/tasks", "application/json", strings.NewReader(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	return res, raw
}

// mustCreateChecklistTask creates a task titled title for checklist contract tests.
func mustCreateChecklistTask(t *testing.T, srv *httptest.Server, title string) string {
	t.Helper()
	return mustCreateTask(t, srv.URL, `{"title":"`+title+`","priority":"medium"}`)
}
