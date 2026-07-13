package handler

import (
	"context"
	"encoding/json"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestHTTP_deleteChecklistItem_204ThenGone(t *testing.T) {
	srv, st := newTaskCreateTestServerWithStore(t)
	defer srv.Close()
	taskID := mustCreateChecklistTask(t, srv, "chk-del")
	it, err := st.AddChecklistItem(context.Background(), taskID, "remove me", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodDelete,
		srv.URL+"/tasks/"+taskID+"/checklist/items/"+it.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d (want 204) body=%s", res.StatusCode, body)
	}
	if len(body) != 0 {
		t.Fatalf("DELETE 204 body must be empty, got %q", body)
	}

	getRes, err := http.Get(srv.URL + "/tasks/" + taskID + "/checklist")
	if err != nil {
		t.Fatal(err)
	}
	defer getRes.Body.Close()
	if getRes.StatusCode != http.StatusOK {
		t.Fatalf("GET checklist status %d", getRes.StatusCode)
	}
	var got struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(getRes.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	for _, item := range got.Items {
		if item.ID == it.ID {
			t.Fatalf("deleted item %q still present in GET checklist", it.ID)
		}
	}
}

// TestHTTP_checklist_404OnUnknownTask pins the documented 404 mapping across
// the three checklist write routes when the task id does not exist.
func TestHTTP_deleteChecklistItem_publishesTaskUpdated(t *testing.T) {
	srv, st, hub := newSSETriggerServer(t)
	defer srv.Close()
	taskID := mustCreateChecklistTask(t, srv, "chk-sse-del")
	it, err := st.AddChecklistItem(context.Background(), taskID, "remove", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	ch, unsub := hub.Subscribe()
	defer unsub()

	req, err := http.NewRequest(http.MethodDelete,
		srv.URL+"/tasks/"+taskID+"/checklist/items/"+it.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d (want 204) body=%s", res.StatusCode, raw)
	}

	events := drainSSE(t, ch, 1, 2*time.Second)
	got := summarize(events)
	mustEqualEvents(t, "DELETE /tasks/{id}/checklist/items/{itemId}", got, []string{"task_updated:" + taskID})
	mustHaveTaskUpdatedData(t, "DELETE /tasks/{id}/checklist/items/{itemId}", events, taskID)
}

// TestHTTP_deleteChecklistItem_errorPathsNeverPublish pins the negative-side
// SSE invariant: 400 (inheriting child) and 404 (unknown task, unknown item)
// must never publish.
func TestHTTP_deleteChecklistItem_errorPathsNeverPublish(t *testing.T) {
	srv, st, hub := newSSETriggerServer(t)
	defer srv.Close()
	parentID := mustCreateChecklistTask(t, srv, "chk-sse-del-par")
	it, err := st.AddChecklistItem(context.Background(), parentID, "owned", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	ch, unsub := hub.Subscribe()
	defer unsub()

	doDelete := func(name, taskID, itemID string, want int) {
		t.Helper()
		req, err := http.NewRequest(http.MethodDelete,
			srv.URL+"/tasks/"+taskID+"/checklist/items/"+itemID, nil)
		if err != nil {
			t.Fatal(err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		raw, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode != want {
			t.Fatalf("%s: status %d (want %d) body=%s", name, res.StatusCode, want, raw)
		}
	}

	doDelete("unknownTask", "11111111-1111-4111-8111-111111111111", it.ID, http.StatusNotFound)
	doDelete("unknownItem", parentID, "22222222-2222-4222-8222-222222222222", http.StatusNotFound)

	got := summarize(drainSSE(t, ch, 0, 200*time.Millisecond))
	if len(got) != 0 {
		t.Fatalf("drained SSE events %v after DELETE checklist error round-trips; want zero (400/404 paths must never publish)", got)
	}
}
