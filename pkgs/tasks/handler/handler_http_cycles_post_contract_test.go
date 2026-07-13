package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func mustStartCycle(t *testing.T, baseURL, taskID string) string {
	t.Helper()
	res, raw := doCyclesRequest(t, http.MethodPost, baseURL+"/tasks/"+taskID+"/cycles", `{}`)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("seed cycle: status %d body=%s", res.StatusCode, raw)
	}
	var c taskCycleResponse
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("decode seed cycle: %v body=%s", err, raw)
	}
	return c.ID
}

// TestHTTP_postTaskCycle_400ErrorStrings pins every documented 400 string
// for POST /tasks/{id}/cycles. Drift in any of these phrases breaks the test
// in lockstep with docs/api.md (Stage 6 commits the doc rows).
func TestHTTP_postTaskCycle_400ErrorStrings(t *testing.T) {
	srv := newTaskCreateTestServer(t)
	defer srv.Close()
	taskID := mustCreateTaskForCycles(t, srv.URL)

	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "unknownField",
			body: `{"nope":1}`,
			want: `json: unknown field "nope"`,
		},
		{
			name: "trailingData",
			body: `{}{}`,
			want: "request body must contain a single JSON value",
		},
		{
			name: "emptyParentString",
			body: `{"parent_cycle_id":""}`,
			want: "parent_cycle_id",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, raw := doCyclesRequest(t, http.MethodPost, srv.URL+"/tasks/"+taskID+"/cycles", tc.body)
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status %d (want 400) body=%s", res.StatusCode, raw)
			}
			var errBody jsonErrorBody
			if err := json.Unmarshal(raw, &errBody); err != nil {
				t.Fatalf("decode: %v body=%s", err, raw)
			}
			if errBody.Error != tc.want {
				t.Fatalf("error=%q want %q", errBody.Error, tc.want)
			}
		})
	}
}

// TestHTTP_postTaskCycle_rejectsConcurrentRunning ensures the "at most one
// running cycle per task" invariant from the store surfaces as a 400 with
// the documented bare phrase, not a 500.
func TestHTTP_postTaskCycle_rejectsConcurrentRunning(t *testing.T) {
	srv := newTaskCreateTestServer(t)
	defer srv.Close()
	taskID := mustCreateTaskForCycles(t, srv.URL)
	_ = mustStartCycle(t, srv.URL, taskID)

	res, raw := doCyclesRequest(t, http.MethodPost, srv.URL+"/tasks/"+taskID+"/cycles", `{}`)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d (want 400) body=%s", res.StatusCode, raw)
	}
	var errBody jsonErrorBody
	if err := json.Unmarshal(raw, &errBody); err != nil {
		t.Fatal(err)
	}
	if errBody.Error != "task already has a running cycle" {
		t.Fatalf("error=%q", errBody.Error)
	}
}

// TestHTTP_postTaskCycle_taskNotFound_returns_404 exercises the FK-style
// 404 path: nonexistent task → ErrNotFound mapped to 404 with the standard
// "not found" body.
func TestHTTP_postTaskCycle_taskNotFound_returns_404(t *testing.T) {
	srv := newTaskCreateTestServer(t)
	defer srv.Close()

	res, raw := doCyclesRequest(t, http.MethodPost,
		srv.URL+"/tasks/00000000-0000-0000-0000-000000000099/cycles", `{}`)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d body=%s", res.StatusCode, raw)
	}
	var errBody jsonErrorBody
	if err := json.Unmarshal(raw, &errBody); err != nil {
		t.Fatal(err)
	}
	if errBody.Error != "not found" {
		t.Fatalf("error=%q", errBody.Error)
	}
}

// TestHTTP_postTaskCycle_response_shape pins the JSON envelope for the
// 201 success path. Adding or removing a top-level key without updating
// the docs (and the web client) breaks here.
func TestHTTP_postTaskCycle_response_shape(t *testing.T) {
	srv := newTaskCreateTestServer(t)
	defer srv.Close()
	taskID := mustCreateTaskForCycles(t, srv.URL)

	res, raw := doCyclesRequest(t, http.MethodPost, srv.URL+"/tasks/"+taskID+"/cycles", `{}`)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d body=%s", res.StatusCode, raw)
	}
	if got := res.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("content-type=%q", got)
	}

	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode: %v body=%s", err, raw)
	}
	requireKeys := []string{"id", "task_id", "attempt_seq", "status", "started_at", "triggered_by", "meta"}
	for _, k := range requireKeys {
		if _, ok := doc[k]; !ok {
			t.Fatalf("missing key %q (full=%v)", k, doc)
		}
	}
	for _, k := range []string{"ended_at", "parent_cycle_id"} {
		if _, ok := doc[k]; ok {
			t.Fatalf("optional key %q must be omitted on a fresh running cycle (full=%v)", k, doc)
		}
	}
}

// TestHTTP_getTaskCycles_400ErrorStrings pins limit-validation messages.
func TestHTTP_postTaskCyclePhase_400ErrorStrings(t *testing.T) {
	srv := newTaskCreateTestServer(t)
	defer srv.Close()
	taskID := mustCreateTaskForCycles(t, srv.URL)
	cycleID := mustStartCycle(t, srv.URL, taskID)

	cases := []struct {
		name string
		body string
		want string
	}{
		{"unknownField", `{"phase":"execute","oops":1}`, `json: unknown field "oops"`},
		{"emptyPhase", `{}`, "phase"},
		{"invalidPhase", `{"phase":"nope"}`, "phase"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, raw := doCyclesRequest(t, http.MethodPost,
				srv.URL+"/tasks/"+taskID+"/cycles/"+cycleID+"/phases", tc.body)
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status %d body=%s", res.StatusCode, raw)
			}
			var errBody jsonErrorBody
			if err := json.Unmarshal(raw, &errBody); err != nil {
				t.Fatal(err)
			}
			if errBody.Error != tc.want {
				t.Fatalf("error=%q want %q", errBody.Error, tc.want)
			}
		})
	}
}

// TestHTTP_postTaskCyclePhase_invalid_transition_400 — phase state machine
// rejection from the store reaches HTTP unchanged.
func TestHTTP_postTaskCyclePhase_invalid_transition_400(t *testing.T) {
	srv := newTaskCreateTestServer(t)
	defer srv.Close()
	taskID := mustCreateTaskForCycles(t, srv.URL)
	cycleID := mustStartCycle(t, srv.URL, taskID)

	res, raw := doCyclesRequest(t, http.MethodPost,
		srv.URL+"/tasks/"+taskID+"/cycles/"+cycleID+"/phases", `{"phase":"verify"}`)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d body=%s", res.StatusCode, raw)
	}
	var errBody jsonErrorBody
	if err := json.Unmarshal(raw, &errBody); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(errBody.Error, "phase transition") {
		t.Fatalf("error=%q must start with 'phase transition'", errBody.Error)
	}
}

// TestHTTP_postTaskCyclePhase_404_when_taskMismatch — same cross-task guard.
func TestHTTP_postTaskCyclePhase_404_when_taskMismatch(t *testing.T) {
	srv := newTaskCreateTestServer(t)
	defer srv.Close()
	taskA := mustCreateTaskForCycles(t, srv.URL)
	taskB := mustCreateTaskForCycles(t, srv.URL)
	cycleID := mustStartCycle(t, srv.URL, taskA)

	res, raw := doCyclesRequest(t, http.MethodPost,
		srv.URL+"/tasks/"+taskB+"/cycles/"+cycleID+"/phases", `{"phase":"execute"}`)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d body=%s", res.StatusCode, raw)
	}
}

// TestHTTP_patchTaskCyclePhase_400_path_validation pins {phaseSeq} parsing.
