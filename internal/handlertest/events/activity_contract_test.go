package events_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// TestHTTP_taskActivity_responseShape pins the documented GET /tasks/activity
// response shape from docs/api.md: total, limit, offset, events[]. Each event
// row must carry task_id, seq, at, type, by, data (defaulted to {} when no
// payload). Cache-Control must be no-store (WriteJSON path).
func TestHTTP_taskActivity_responseShape(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()

	task := handlertest.PostTaskJSON(t, srv, `{"title":"activity-shape","priority":"medium"}`, http.StatusCreated)
	if err := st.AppendTaskEvent(context.Background(), task.ID, taskeventsdomain.EventStatusChanged, taskcoredomain.ActorUser, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}

	res, err := http.Get(srv.URL + "/tasks/activity")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d body=%s", res.StatusCode, body)
	}

	if got := res.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q want no-store (docs/api.md)", got)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decode: %v body=%s", err, body)
	}
	for _, k := range []string{"total", "limit", "offset", "events"} {
		if _, ok := raw[k]; !ok {
			t.Errorf("response missing required key %q (docs/api.md): %s", k, body)
		}
	}

	var events []map[string]json.RawMessage
	if err := json.Unmarshal(raw["events"], &events); err != nil {
		t.Fatalf("decode events: %v", err)
	}
	if len(events) < 1 {
		t.Fatalf("want >=1 event row (status_changed appended), got 0")
	}
	for i, ev := range events {
		for _, k := range []string{"task_id", "seq", "at", "type", "by", "data"} {
			if _, ok := ev[k]; !ok {
				t.Errorf("event[%d] missing required key %q: %v", i, k, ev)
			}
		}
		if string(ev["data"]) == "" || string(ev["data"]) == "null" {
			t.Errorf("event[%d] data must not be null/empty (defaults to {}): got %q", i, ev["data"])
		}
	}

	var limit int
	if err := json.Unmarshal(raw["limit"], &limit); err != nil || limit != 50 {
		t.Errorf("default limit want 50, got raw=%s err=%v", raw["limit"], err)
	}
	var offset int
	if err := json.Unmarshal(raw["offset"], &offset); err != nil || offset != 0 {
		t.Errorf("default offset want 0, got raw=%s err=%v", raw["offset"], err)
	}
}

// TestHTTP_taskActivity_typeFilter ensures only the three types
// (status_changed, phase_failed, approval_granted) appear in the feed;
// task_created events must not appear.
func TestHTTP_taskActivity_typeFilter(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()

	task := handlertest.PostTaskJSON(t, srv, `{"title":"activity-filter","priority":"medium"}`, http.StatusCreated)
	if err := st.AppendTaskEvent(context.Background(), task.ID, taskeventsdomain.EventStatusChanged, taskcoredomain.ActorAgent, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if err := st.AppendTaskEvent(context.Background(), task.ID, taskeventsdomain.EventApprovalGranted, taskcoredomain.ActorUser, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}

	body, _ := handlertest.MustGetJSON(t, srv.URL, "/tasks/activity")
	var resp struct {
		Events []struct {
			Type string `json:"type"`
		} `json:"events"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	allowed := map[string]bool{
		"status_changed":  true,
		"phase_failed":    true,
		"approval_granted": true,
	}
	for i, ev := range resp.Events {
		if !allowed[ev.Type] {
			t.Errorf("event[%d] type=%q must not appear in /tasks/activity", i, ev.Type)
		}
	}
	var foundStatus, foundApproval bool
	for _, ev := range resp.Events {
		if ev.Type == "status_changed" {
			foundStatus = true
		}
		if ev.Type == "approval_granted" {
			foundApproval = true
		}
	}
	if !foundStatus || !foundApproval {
		t.Errorf("want status_changed and approval_granted in feed, got types: %v", resp.Events)
	}
}

// TestHTTP_taskActivity_newestFirst ensures ORDER BY at DESC, seq DESC.
func TestHTTP_taskActivity_newestFirst(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()

	task := handlertest.PostTaskJSON(t, srv, `{"title":"activity-order","priority":"medium"}`, http.StatusCreated)
	for i := 0; i < 3; i++ {
		if err := st.AppendTaskEvent(context.Background(), task.ID, taskeventsdomain.EventStatusChanged, taskcoredomain.ActorUser, []byte(`{}`)); err != nil {
			t.Fatal(err)
		}
	}

	body, _ := handlertest.MustGetJSON(t, srv.URL, "/tasks/activity?limit=10")
	var resp struct {
		Events []struct {
			Seq int64  `json:"seq"`
			At  string `json:"at"`
		} `json:"events"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Events) < 2 {
		t.Skip("not enough events to check ordering")
	}
	for i := 1; i < len(resp.Events); i++ {
		t0, _ := time.Parse(time.RFC3339, resp.Events[i-1].At)
		t1, _ := time.Parse(time.RFC3339, resp.Events[i].At)
		if t0.Before(t1) {
			t.Errorf("events[%d] at=%s is before events[%d] at=%s — want newest first",
				i-1, resp.Events[i-1].At, i, resp.Events[i].At)
		}
	}
}

// TestHTTP_taskActivity_pagination tests limit/offset pagination.
func TestHTTP_taskActivity_pagination(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()

	task := handlertest.PostTaskJSON(t, srv, `{"title":"activity-page","priority":"medium"}`, http.StatusCreated)
	for i := 0; i < 5; i++ {
		if err := st.AppendTaskEvent(context.Background(), task.ID, taskeventsdomain.EventStatusChanged, taskcoredomain.ActorUser, []byte(`{}`)); err != nil {
			t.Fatal(err)
		}
	}

	body1, _ := handlertest.MustGetJSON(t, srv.URL, "/tasks/activity?limit=2&offset=0")
	var page1 struct {
		Total  int64 `json:"total"`
		Limit  int   `json:"limit"`
		Offset int   `json:"offset"`
		Events []struct {
			Seq int64 `json:"seq"`
		} `json:"events"`
	}
	if err := json.Unmarshal(body1, &page1); err != nil {
		t.Fatalf("decode page1: %v", err)
	}
	if page1.Limit != 2 {
		t.Errorf("limit want 2, got %d", page1.Limit)
	}
	if page1.Offset != 0 {
		t.Errorf("offset want 0, got %d", page1.Offset)
	}
	if len(page1.Events) != 2 {
		t.Errorf("want 2 events on page1, got %d", len(page1.Events))
	}
	if page1.Total < 5 {
		t.Errorf("total want >=5, got %d", page1.Total)
	}

	body2, _ := handlertest.MustGetJSON(t, srv.URL, "/tasks/activity?limit=2&offset=2")
	var page2 struct {
		Events []struct {
			Seq int64 `json:"seq"`
		} `json:"events"`
	}
	if err := json.Unmarshal(body2, &page2); err != nil {
		t.Fatalf("decode page2: %v", err)
	}
	if len(page1.Events) > 0 && len(page2.Events) > 0 {
		if page1.Events[0].Seq == page2.Events[0].Seq {
			t.Errorf("page1 and page2 share first seq %d — offset not applied", page1.Events[0].Seq)
		}
	}
}

// TestHTTP_taskActivity_sinceFilter tests the since RFC3339 lower bound.
func TestHTTP_taskActivity_sinceFilter(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()

	task := handlertest.PostTaskJSON(t, srv, `{"title":"activity-since","priority":"medium"}`, http.StatusCreated)
	if err := st.AppendTaskEvent(context.Background(), task.ID, taskeventsdomain.EventStatusChanged, taskcoredomain.ActorUser, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}

	futureTime := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	body, _ := handlertest.MustGetJSON(t, srv.URL, "/tasks/activity?since="+futureTime)
	var resp struct {
		Events []json.RawMessage `json:"events"`
		Total  int64             `json:"total"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Events) != 0 {
		t.Errorf("since=future want 0 events, got %d", len(resp.Events))
	}
	if resp.Total != 0 {
		t.Errorf("since=future want total=0, got %d", resp.Total)
	}
}

// TestHTTP_taskActivity_validation400s pins every documented 400 error string.
func TestHTTP_taskActivity_validation400s(t *testing.T) {
	srv, _ := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()

	long := strings.Repeat("9", 33)
	cases := []struct {
		name, url, wantSubstr string
	}{
		{"limit too long", "/tasks/activity?limit=" + long, "limit value too long"},
		{"limit zero", "/tasks/activity?limit=0", "limit must be integer 1..200"},
		{"limit over max", "/tasks/activity?limit=201", "limit must be integer 1..200"},
		{"limit non-numeric", "/tasks/activity?limit=nope", "limit must be integer 1..200"},
		{"offset too long", "/tasks/activity?offset=" + long, "offset value too long"},
		{"offset negative", "/tasks/activity?offset=-1", "offset must be non-negative integer"},
		{"offset non-numeric", "/tasks/activity?offset=nope", "offset must be non-negative integer"},
		{"since too long", "/tasks/activity?since=" + strings.Repeat("x", 65), "since value too long"},
		{"since invalid", "/tasks/activity?since=not-a-date", "since must be RFC3339 timestamp"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := http.Get(srv.URL + tc.url)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status %d (want 400) body=%s", res.StatusCode, body)
			}
			if !strings.Contains(string(body), tc.wantSubstr) {
				t.Fatalf("error body must contain %q (docs/api.md), got %s", tc.wantSubstr, body)
			}
		})
	}
}

// TestHTTP_taskActivity_emptyFeed ensures a valid 200 with empty events array
// when no matching events exist.
func TestHTTP_taskActivity_emptyFeed(t *testing.T) {
	srv, _ := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()

	body, res := handlertest.MustGetJSON(t, srv.URL, "/tasks/activity")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var resp struct {
		Total  int64             `json:"total"`
		Events []json.RawMessage `json:"events"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Events == nil {
		t.Error("events must be [] not null when empty")
	}
}

// TestHTTP_taskActivity_taskTitlePresent verifies optional task_title field
// is included when the task exists.
func TestHTTP_taskActivity_taskTitlePresent(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()

	const title = "activity-title-task"
	task := handlertest.PostTaskJSON(t, srv, `{"title":"`+title+`","priority":"medium"}`, http.StatusCreated)
	if err := st.AppendTaskEvent(context.Background(), task.ID, taskeventsdomain.EventStatusChanged, taskcoredomain.ActorUser, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}

	body, _ := handlertest.MustGetJSON(t, srv.URL, "/tasks/activity")
	var resp struct {
		Events []struct {
			TaskID    string  `json:"task_id"`
			TaskTitle *string `json:"task_title"`
		} `json:"events"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, ev := range resp.Events {
		if ev.TaskID != task.ID {
			continue
		}
		if ev.TaskTitle == nil || *ev.TaskTitle != title {
			t.Errorf("task_title want %q, got %v", title, ev.TaskTitle)
		}
		return
	}
	t.Errorf("no event for task %s found in activity feed", task.ID)
}
