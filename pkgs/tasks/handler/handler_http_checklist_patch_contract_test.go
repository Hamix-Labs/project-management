package handler

import (
	"context"
	"encoding/json"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHTTP_patchChecklistItem_doneAgentReturnsItemsView(t *testing.T) {
	srv, st := newTaskCreateTestServerWithStore(t)
	defer srv.Close()
	taskID := mustCreateChecklistTask(t, srv, "chk-done")
	it, err := st.AddChecklistItem(context.Background(), taskID, "review", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPatch,
		srv.URL+"/tasks/"+taskID+"/checklist/items/"+it.ID,
		strings.NewReader(`{"done":true,"evidence":"reviewed in test","verified_by":"agent_self"}`))
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
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d (want 200) body=%s", res.StatusCode, raw)
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatalf("decode: %v body=%s", err, raw)
	}
	if _, ok := top["items"]; !ok || len(top) != 1 {
		t.Fatalf("PATCH 200 must return only `items`; got keys=%v body=%s", keysOf(top), raw)
	}
	var out struct {
		Items []struct {
			ID         string `json:"id"`
			SortOrder  int    `json:"sort_order"`
			Text       string `json:"text"`
			Done       bool   `json:"done"`
			Evidence   string `json:"evidence"`
			VerifiedBy string `json:"verified_by"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode items: %v", err)
	}
	if len(out.Items) != 2 {
		t.Fatalf("items len=%d want 2", len(out.Items))
	}
	var got *struct {
		ID         string `json:"id"`
		SortOrder  int    `json:"sort_order"`
		Text       string `json:"text"`
		Done       bool   `json:"done"`
		Evidence   string `json:"evidence"`
		VerifiedBy string `json:"verified_by"`
	}
	for i := range out.Items {
		if out.Items[i].ID == it.ID {
			got = &out.Items[i]
			break
		}
	}
	if got == nil {
		t.Fatalf("patched item missing from %#v", out.Items)
	}
	if got.Text != "review" || !got.Done {
		t.Fatalf("item=%+v want id=%s text=review done=true", got, it.ID)
	}
}

// TestHTTP_deleteChecklistItem_204ThenGone pins the documented DELETE happy
// path: 204 with empty body and the row vanishing from a follow-up GET.
func TestHTTP_patchChecklistItem_textBranch400Strings(t *testing.T) {
	srv := newTaskCreateTestServer(t)
	defer srv.Close()

	parentID := mustCreateChecklistTask(t, srv, "chk-text-400-parent")
	defRes, err := http.Post(srv.URL+"/tasks/"+parentID+"/checklist/items",
		"application/json", strings.NewReader(`{"text":"owned"}`))
	if err != nil {
		t.Fatal(err)
	}
	defBody, _ := io.ReadAll(defRes.Body)
	_ = defRes.Body.Close()
	if defRes.StatusCode != http.StatusCreated {
		t.Fatalf("seed item status %d body=%s", defRes.StatusCode, defBody)
	}
	var def checklistdomain.TaskChecklistItem
	if err := json.Unmarshal(defBody, &def); err != nil {
		t.Fatal(err)
	}

	patch := func(taskID, itemID, body string) (int, string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodPatch,
			srv.URL+"/tasks/"+taskID+"/checklist/items/"+itemID,
			strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		return res.StatusCode, string(raw)
	}

	cases := []struct {
		name             string
		taskID, itemID   string
		body             string
		want             string
		commentaryReason string
	}{
		{
			name: "emptyText", taskID: parentID, itemID: def.ID,
			body:             `{"text":""}`,
			want:             "text required",
			commentaryReason: "handler-generated phrase before the store call (handler_checklist.go:103); empty/whitespace-only text is rejected here so the store never sees it",
		},
		{
			name: "whitespaceText", taskID: parentID, itemID: def.ID,
			body:             `{"text":"   \t  "}`,
			want:             "text required",
			commentaryReason: "same handler branch as emptyText; trim happens before the empty check so all-whitespace produces the same wire phrase rather than being silently dropped to a store-side error",
		},
		{
			name: "noFields", taskID: parentID, itemID: def.ID,
			body:             `{}`,
			want:             "send exactly one of text, verify_commands, or done",
			commentaryReason: "shared one-of-choice phrase that gates BOTH text and done branches (handler_checklist.go:95); the existing errorPathsNeverPublish test only checks status code, so the bare phrase needs its own pin",
		},
		{
			name: "bothFields", taskID: parentID, itemID: def.ID,
			body:             `{"text":"x","done":true}`,
			want:             "send exactly one of text, verify_commands, or done",
			commentaryReason: "same one-of phrase from the opposite direction (sending both fields); proves the textSet == doneSet branch covers the symmetric case the doc bullet `or neither field was provided for the one-of choice` only covers the empty side of",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, raw := patch(tc.taskID, tc.itemID, tc.body)
			if code != http.StatusBadRequest {
				t.Fatalf("status %d (want 400) body=%s — case rationale: %s", code, raw, tc.commentaryReason)
			}
			var errBody jsonErrorBody
			if err := json.Unmarshal([]byte(raw), &errBody); err != nil {
				t.Fatalf("decode: %v body=%s", err, raw)
			}
			if errBody.Error != tc.want {
				t.Fatalf("error=%q want %q (docs/api.md PATCH /tasks/{id}/checklist/items/{itemId} 400 strings) — case rationale: %s", errBody.Error, tc.want, tc.commentaryReason)
			}
		})
	}
}

// TestHTTP_patchChecklistItem_publishesTaskUpdated colocates the per-route SSE
// positive invariant for PATCH: a successful done-toggle publishes exactly
// `task_updated:{id}`.
func TestHTTP_patchChecklistItem_publishesTaskUpdated(t *testing.T) {
	srv, st, hub := newSSETriggerServer(t)
	defer srv.Close()
	taskID := mustCreateChecklistTask(t, srv, "chk-sse-patch")
	it, err := st.AddChecklistItem(context.Background(), taskID, "review", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	ch, unsub := hub.Subscribe()
	defer unsub()

	req, err := http.NewRequest(http.MethodPatch,
		srv.URL+"/tasks/"+taskID+"/checklist/items/"+it.ID,
		strings.NewReader(`{"done":true,"evidence":"sse test","verified_by":"agent_self"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor", "agent")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d (want 200) body=%s", res.StatusCode, raw)
	}

	events := drainSSE(t, ch, 1, 2*time.Second)
	got := summarize(events)
	mustEqualEvents(t, "PATCH /tasks/{id}/checklist/items/{itemId}", got, []string{"task_updated:" + taskID})
	mustHaveTaskUpdatedData(t, "PATCH /tasks/{id}/checklist/items/{itemId}", events, taskID)
}

// TestHTTP_patchChecklistItem_errorPathsNeverPublish pins the negative-side SSE
// invariant: 400 (done by user actor, unknown done value) and 404 (unknown
// task, unknown item) must never publish.
func TestHTTP_patchChecklistItem_errorPathsNeverPublish(t *testing.T) {
	srv, st, hub := newSSETriggerServer(t)
	defer srv.Close()
	taskID := mustCreateChecklistTask(t, srv, "chk-sse-patch-neg")
	it, err := st.AddChecklistItem(context.Background(), taskID, "neg", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	ch, unsub := hub.Subscribe()
	defer unsub()

	doPatch := func(name, taskID, itemID, body, actor string, want int) {
		t.Helper()
		req, err := http.NewRequest(http.MethodPatch,
			srv.URL+"/tasks/"+taskID+"/checklist/items/"+itemID,
			strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		if actor != "" {
			req.Header.Set("X-Actor", actor)
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

	doPatch("doneByUser", taskID, it.ID, `{"done":true}`, "user", http.StatusBadRequest)
	doPatch("noFields", taskID, it.ID, `{}`, "agent", http.StatusBadRequest)
	doPatch("unknownTask", "11111111-1111-4111-8111-111111111111", it.ID, `{"text":"x"}`, "", http.StatusNotFound)
	doPatch("unknownItem", taskID, "22222222-2222-4222-8222-222222222222", `{"text":"x"}`, "", http.StatusNotFound)

	got := summarize(drainSSE(t, ch, 0, 200*time.Millisecond))
	if len(got) != 0 {
		t.Fatalf("drained SSE events %v after PATCH checklist error round-trips; want zero (400/404 paths must never publish)", got)
	}
}

// TestHTTP_deleteChecklistItem_publishesTaskUpdated colocates the per-route SSE
// positive invariant for DELETE: a successful removal publishes exactly
// `task_updated:{id}`.
func keysOf(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
