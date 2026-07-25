package taskcore_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func postClose(t *testing.T, baseURL, id string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, baseURL+"/tasks/"+id+"/close", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	return res, raw
}

func postReopen(t *testing.T, baseURL, id string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, baseURL+"/tasks/"+id+"/reopen", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	return res, raw
}

func TestHTTP_closeTask_setsClosed(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	id := handlertest.MustCreateTask(t, srv.URL, `{"title":"to-close","priority":"medium"}`)
	res, raw := postClose(t, srv.URL, id)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d body=%s", res.StatusCode, raw)
	}
	var task taskcoredomain.Task
	if err := json.Unmarshal(raw, &task); err != nil {
		t.Fatal(err)
	}
	if task.Status != taskcoredomain.StatusClosed {
		t.Fatalf("status=%q want closed", task.Status)
	}

	// Idempotent.
	res2, raw2 := postClose(t, srv.URL, id)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("reclose status %d body=%s", res2.StatusCode, raw2)
	}
}

func TestHTTP_reopenTask_closedToReady(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	id := handlertest.MustCreateTask(t, srv.URL, `{"title":"to-reopen","priority":"medium"}`)
	if res, raw := postClose(t, srv.URL, id); res.StatusCode != http.StatusOK {
		t.Fatalf("close status %d body=%s", res.StatusCode, raw)
	}
	res, raw := postReopen(t, srv.URL, id)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("reopen status %d body=%s", res.StatusCode, raw)
	}
	var task taskcoredomain.Task
	if err := json.Unmarshal(raw, &task); err != nil {
		t.Fatal(err)
	}
	if task.Status != taskcoredomain.StatusReady {
		t.Fatalf("status=%q want ready", task.Status)
	}
}

func TestHTTP_reopenTask_notClosedIs409(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	id := handlertest.MustCreateTask(t, srv.URL, `{"title":"open","priority":"medium"}`)
	res, raw := postReopen(t, srv.URL, id)
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("status %d want 409 body=%s", res.StatusCode, raw)
	}
}

func TestHTTP_patchStatusClosedRejected(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	id := handlertest.MustCreateTask(t, srv.URL, `{"title":"patch-closed","priority":"medium"}`)
	res, raw := handlertest.PatchTask(t, srv.URL, id, `{"status":"closed"}`)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d want 400 body=%s", res.StatusCode, raw)
	}
}

func TestHTTP_deleteTask_routeGone(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	id := handlertest.MustCreateTask(t, srv.URL, `{"title":"no-delete","priority":"medium"}`)
	res, raw := handlertest.DeleteTask(t, srv.URL, id)
	if res.StatusCode != http.StatusMethodNotAllowed && res.StatusCode != http.StatusNotFound {
		t.Fatalf("DELETE status %d want 404 or 405 body=%s", res.StatusCode, raw)
	}
}

func TestHTTP_closeTask_publishesTaskUpdated(t *testing.T) {
	srv, _, hub := handlertest.NewSSETriggerServer(t)
	defer srv.Close()

	id := handlertest.MustCreateTask(t, srv.URL, `{"title":"close-sse","priority":"medium"}`)
	ch, unsub := hub.Subscribe()
	defer unsub()

	if res, raw := postClose(t, srv.URL, id); res.StatusCode != http.StatusOK {
		t.Fatalf("close status %d body=%s", res.StatusCode, raw)
	}
	got := handlertest.SummarizeSSEEvents(handlertest.DrainSSE(t, ch, 1, 2*time.Second))
	handlertest.MustEqualEvents(t, "POST /tasks/{id}/close", got, []string{string(realtime.TaskUpdated) + ":" + id})
}
