package taskcore_test

import (
	"encoding/json"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/storefake"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

func TestHTTP_getTask_flatTaskEnvelope(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	id := handlertest.MustCreateTask(t, srv.URL, `{"title":"root-leaf","priority":"medium"}`)

	deadline := time.Now().Add(15 * time.Second)
	var raw []byte
	var res *http.Response
	for {
		res, raw = handlertest.GetTask(t, srv.URL, id)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status %d (want 200) body=%s", res.StatusCode, raw)
		}
		if strings.Contains(string(raw), `"worktree_id"`) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for worktree_id in body=%s", raw)
		}
		time.Sleep(25 * time.Millisecond)
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, raw)
	}

	wantKeys := []string{"created_at", "cursor_model", "id", "initial_prompt", "number", "pickup_not_before", "priority", "project_id", "runner", "runner_config", "status", "title", "verify_chat_mode", "worktree_id"}
	gotKeys := make([]string, 0, len(top))
	for k := range top {
		gotKeys = append(gotKeys, k)
	}
	sort.Strings(gotKeys)
	sort.Strings(wantKeys)
	if !handlertest.EqualStringSlices(gotKeys, wantKeys) {
		t.Fatalf("flat task envelope keys=%v want %v", gotKeys, wantKeys)
	}
	if strings.Contains(string(raw), `"children"`) {
		t.Fatalf("body=%s contains \"children\" key (flat task response)", raw)
	}
}

func TestHTTP_getTask_pathSegmentGuard(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	cases := []struct {
		name string
		path string
		want string
	}{
		{"whitespaceOnlyID", "%20%20%20", "id"},
		{"overlongID", strings.Repeat("a", handlerhttp.MaxPathIDBytes+1), "id too long"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, raw := handlertest.GetTask(t, srv.URL, tc.path)
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status %d (want 400) body=%s", res.StatusCode, raw)
			}
			var errBody handlertest.JSONErrorBody
			if err := json.Unmarshal(raw, &errBody); err != nil {
				t.Fatalf("decode: %v body=%s", err, raw)
			}
			if errBody.Error != tc.want {
				t.Fatalf("error=%q want %q", errBody.Error, tc.want)
			}
		})
	}
}

func TestHTTP_getTask_unknownIDIs404(t *testing.T) {
	st := storefake.NewHandlerStore()
	st.FailGet(taskcoredomain.ErrNotFound)
	srv := httptest.NewServer(handler.NewHandler(st, realtime.NewSSEHub(), nil))
	t.Cleanup(srv.Close)

	taskID := "11111111-1111-4111-8111-111111111111"
	res, raw := handlertest.GetTask(t, srv.URL, taskID)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d (want 404) body=%s", res.StatusCode, raw)
	}
	var errBody handlertest.JSONErrorBody
	if err := json.Unmarshal(raw, &errBody); err != nil {
		t.Fatalf("decode: %v body=%s", err, raw)
	}
	if errBody.Error != "not found" {
		t.Fatalf("error=%q want %q", errBody.Error, "not found")
	}
	calls := st.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("Get calls=%d want 1", len(calls))
	}
	if calls[0].ID != taskID {
		t.Fatalf("Get id=%q want %q", calls[0].ID, taskID)
	}
}

func TestHTTP_getTask_trailingSlashIsMux404(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/tasks/")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d (want 404 from mux) body=%s", res.StatusCode, raw)
	}
	if strings.Contains(string(raw), `"tasks":`) {
		t.Fatalf("body=%s looks like the GET /tasks list envelope", raw)
	}
	if strings.Contains(string(raw), `"error"`) {
		t.Fatalf("body=%s contains JSON error envelope; mux 404 must produce text body only", raw)
	}
}

func TestHTTP_getTask_neverPublishesOnSSE(t *testing.T) {
	srv, _, hub := handlertest.NewSSETriggerServer(t)
	defer srv.Close()

	id := handlertest.MustCreateTask(t, srv.URL, `{"title":"p","priority":"medium"}`)

	ch, unsub := hub.Subscribe()
	defer unsub()

	if res, raw := handlertest.GetTask(t, srv.URL, id); res.StatusCode != http.StatusOK {
		t.Fatalf("get task status %d body=%s", res.StatusCode, raw)
	}
	res404, raw404 := handlertest.GetTask(t, srv.URL, "11111111-1111-4111-8111-111111111111")
	if res404.StatusCode != http.StatusNotFound {
		t.Fatalf("get unknown status %d body=%s", res404.StatusCode, raw404)
	}

	got := handlertest.SummarizeSSEEvents(handlertest.DrainSSE(t, ch, 0, 200*time.Millisecond))
	if len(got) != 0 {
		t.Fatalf("drained SSE events %v after GET /tasks/{id}; want zero", got)
	}
}
