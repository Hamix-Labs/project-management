package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func newSSETriggerServer(t *testing.T) (*httptest.Server, *composition.API, *SSEHub) {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	st := composition.NewAPI(db)
	binding := seedHandlerTestGitRepo(t, st)
	hub := NewSSEHub()
	srv := httptest.NewServer(newTaskTestHandlerWithHub(st, hub))
	registerHandlerGitBinding(t, srv.URL, binding)
	return srv, st, hub
}

// drainSSE collects up to want events from ch, returning whatever arrived
// within timeout. Returning early when len(out) == want keeps tests fast.
// The caller should assert on the slice rather than block forever.
func drainSSE(t *testing.T, ch <-chan string, want int, timeout time.Duration) []realtime.Event {
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

// summarize collapses a realtime.Event slice into a stable string set so
// tests can compare published events without relying on publish order. The
// format is "type:id" for task-only events, "type:id/cycle_id" for
// task_cycle_changed events, and "type:id/seq" when event_seq is set.
func summarize(events []realtime.Event) []string {
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

func mustEqualEvents(t *testing.T, route string, got, want []string) {
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

func mustHaveTaskUpdatedData(t *testing.T, route string, events []realtime.Event, taskID string) {
	t.Helper()
	for _, ev := range events {
		if ev.Type == realtime.TaskUpdated && ev.ID == taskID {
			if ev.Data == nil {
				t.Fatalf("%s: task_updated:%s missing data enrichment (ADR-0026)", route, taskID)
			}
			return
		}
	}
	t.Fatalf("%s: no task_updated:%s in events %v", route, taskID, summarize(events))
}

// TestHTTP_SSE_triggerSurface pins the SSE trigger table documented in
// docs/api.md. Each subtest exercises one HTTP write and asserts the exact
// set of {type,id} events published on the hub. If a future change adds or
// removes a publish, this test fails so docs/api.md is updated in the same
// PR.
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
	res, err := http.Post(srv.URL+"/tasks", "application/json", strings.NewReader(withCreateChecklistForURL(srv.URL, body)))
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
