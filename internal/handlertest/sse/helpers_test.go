package sse_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func postCycleJSON(t *testing.T, srv *httptest.Server, taskID, body string, wantStatus int) string {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/tasks/"+taskID+"/cycles", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor", "agent")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
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

func postTaskJSON(t *testing.T, srv *httptest.Server, body string, wantStatus int) taskcoredomain.Task {
	t.Helper()
	res, err := http.Post(srv.URL+"/tasks", "application/json", strings.NewReader(handlertest.WithCreateChecklistForURL(srv.URL, body)))
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

func patchTaskJSON(t *testing.T, srv *httptest.Server, id, body string, wantStatus int) {
	t.Helper()
	mustDoJSON(t, http.MethodPatch, srv.URL+"/tasks/"+id, body, "", wantStatus)
}

func mustDoJSON(t *testing.T, method, url, body, actor string, wantStatus int) {
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
