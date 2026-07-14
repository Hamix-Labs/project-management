package taskcore_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

var retryTestHTTPClient = &http.Client{
	Transport: &http.Transport{DisableKeepAlives: true},
}

func postTaskRetry(t *testing.T, baseURL, taskID, body string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, baseURL+"/tasks/"+taskID+"/retry", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	res, err := retryTestHTTPClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return res, raw
}

func retryBody(t *testing.T, mode taskcoredomain.RetryMode, parentCycleID string) string {
	t.Helper()
	payload := map[string]string{"mode": string(mode)}
	if parentCycleID != "" {
		payload["parent_cycle_id"] = parentCycleID
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func mustFailedTaskWithTerminalCycle(t *testing.T, baseURL string) (taskID, cycleID string) {
	t.Helper()
	taskID = handlertest.MustCreateTaskForCycles(t, baseURL)
	_, createdRaw := handlertest.DoCyclesRequest(t, http.MethodPost, baseURL+"/tasks/"+taskID+"/cycles", `{}`)
	var cycle struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createdRaw, &cycle); err != nil {
		t.Fatal(err)
	}
	res, raw := handlertest.DoCyclesRequest(t, http.MethodPatch,
		baseURL+"/tasks/"+taskID+"/cycles/"+cycle.ID, `{"status":"failed","reason":"test failure"}`)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("terminate cycle status %d body=%s", res.StatusCode, raw)
	}
	res2, raw2 := handlertest.PatchTask(t, baseURL, taskID, `{"status":"failed"}`)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("patch task failed status %d body=%s", res2.StatusCode, raw2)
	}
	var got taskcoredomain.Task
	if err := json.Unmarshal(raw2, &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != taskcoredomain.StatusFailed {
		t.Fatalf("task status=%q want failed", got.Status)
	}
	return taskID, cycle.ID
}

func TestHTTP_postTaskRetry_fresh(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	taskID, cycleID := mustFailedTaskWithTerminalCycle(t, srv.URL)

	body := retryBody(t, taskcoredomain.RetryFresh, cycleID)
	res, raw := postTaskRetry(t, srv.URL, taskID, body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d body=%s", res.StatusCode, raw)
	}
	var got taskcoredomain.Task
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != taskcoredomain.StatusReady {
		t.Fatalf("status=%q want ready", got.Status)
	}
	stored, err := st.Get(context.Background(), taskID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.PendingRetry == nil || stored.PendingRetry.Mode != taskcoredomain.RetryFresh || stored.PendingRetry.ParentCycleID != cycleID {
		t.Fatalf("pending_retry=%+v", stored.PendingRetry)
	}
	res2, _ := postTaskRetry(t, srv.URL, taskID, body)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("idempotent status %d", res2.StatusCode)
	}
}

func TestHTTP_postTaskRetry_resume(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	taskID, cycleID := mustFailedTaskWithTerminalCycle(t, srv.URL)

	res, raw := postTaskRetry(t, srv.URL, taskID, retryBody(t, taskcoredomain.RetryResume, cycleID))
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d body=%s", res.StatusCode, raw)
	}
	var got taskcoredomain.Task
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != taskcoredomain.StatusReady {
		t.Fatalf("status=%q want ready", got.Status)
	}
	stored, err := st.Get(context.Background(), taskID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.PendingRetry == nil || stored.PendingRetry.Mode != taskcoredomain.RetryResume {
		t.Fatalf("pending_retry=%+v", stored.PendingRetry)
	}
}

func TestHTTP_postTaskRetry_defaultParentCycle(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	taskID, cycleID := mustFailedTaskWithTerminalCycle(t, srv.URL)

	res, raw := postTaskRetry(t, srv.URL, taskID, retryBody(t, taskcoredomain.RetryFresh, ""))
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d body=%s", res.StatusCode, raw)
	}
	stored, err := st.Get(context.Background(), taskID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.PendingRetry == nil || stored.PendingRetry.ParentCycleID != cycleID {
		t.Fatalf("pending_retry=%+v want parent %s", stored.PendingRetry, cycleID)
	}
}

func TestHTTP_postTaskRetry_rejectsWrongStatus(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskID := handlertest.MustCreateTask(t, srv.URL, `{"title":"ready","priority":"medium","status":"ready"}`)

	res, raw := postTaskRetry(t, srv.URL, taskID, retryBody(t, taskcoredomain.RetryFresh, ""))
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d (want 400) body=%s", res.StatusCode, raw)
	}
}

func TestHTTP_postTaskRetry_rejectsBadParent(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskID, _ := mustFailedTaskWithTerminalCycle(t, srv.URL)

	res, raw := postTaskRetry(t, srv.URL, taskID, retryBody(t, taskcoredomain.RetryFresh, "not-a-cycle"))
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d (want 404) body=%s", res.StatusCode, raw)
	}
}

func TestHTTP_postTaskRetry_conflictDifferentIntent(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskID, cycleID := mustFailedTaskWithTerminalCycle(t, srv.URL)

	freshBody := retryBody(t, taskcoredomain.RetryFresh, cycleID)
	res, _ := postTaskRetry(t, srv.URL, taskID, freshBody)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("first retry status %d", res.StatusCode)
	}
	res2, raw2 := postTaskRetry(t, srv.URL, taskID, retryBody(t, taskcoredomain.RetryResume, cycleID))
	if res2.StatusCode != http.StatusConflict {
		t.Fatalf("status %d (want 409) body=%s", res2.StatusCode, raw2)
	}
}
