package cycles_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
)

func TestHTTP_patchTaskCycle_404_when_taskMismatch(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskA := handlertest.MustCreateTaskForCycles(t, srv.URL)
	taskB := handlertest.MustCreateTaskForCycles(t, srv.URL)
	cycleID := mustStartCycle(t, srv.URL, taskA)

	res, raw := handlertest.DoCyclesRequest(t, http.MethodPatch,
		srv.URL+"/tasks/"+taskB+"/cycles/"+cycleID, `{"status":"succeeded"}`)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("cross-task PATCH cycle: status %d body=%s", res.StatusCode, raw)
	}
}

// TestHTTP_patchTaskCycle_400ErrorStrings pins terminate-status validation.
func TestHTTP_patchTaskCycle_400ErrorStrings(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskID := handlertest.MustCreateTaskForCycles(t, srv.URL)
	cycleID := mustStartCycle(t, srv.URL, taskID)

	cases := []struct {
		name string
		body string
		want string
	}{
		{"unknownField", `{"status":"succeeded","oops":1}`, `json: unknown field "oops"`},
		{"emptyBody", `{}`, "status must be a terminal cycle status"},
		{"runningStatus", `{"status":"running"}`, "status must be a terminal cycle status"},
		{"invalidEnum", `{"status":"nope"}`, "status must be a terminal cycle status"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, raw := handlertest.DoCyclesRequest(t, http.MethodPatch, srv.URL+"/tasks/"+taskID+"/cycles/"+cycleID, tc.body)
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status %d body=%s", res.StatusCode, raw)
			}
			var errBody handlertest.JSONErrorBody
			if err := json.Unmarshal(raw, &errBody); err != nil {
				t.Fatal(err)
			}
			if errBody.Error != tc.want {
				t.Fatalf("error=%q want %q", errBody.Error, tc.want)
			}
		})
	}
}

// TestHTTP_patchTaskCycle_alreadyTerminal_400 — the second terminate is the
// documented 400 path, confirming the store's terminal guard reaches the
// HTTP boundary intact.
func TestHTTP_patchTaskCycle_alreadyTerminal_400(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskID := handlertest.MustCreateTaskForCycles(t, srv.URL)
	cycleID := mustStartCycle(t, srv.URL, taskID)
	if res, raw := handlertest.DoCyclesRequest(t, http.MethodPatch, srv.URL+"/tasks/"+taskID+"/cycles/"+cycleID, `{"status":"succeeded"}`); res.StatusCode != http.StatusOK {
		t.Fatalf("first terminate: %d %s", res.StatusCode, raw)
	}

	res, raw := handlertest.DoCyclesRequest(t, http.MethodPatch, srv.URL+"/tasks/"+taskID+"/cycles/"+cycleID, `{"status":"failed"}`)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("second terminate status %d body=%s", res.StatusCode, raw)
	}
	var errBody handlertest.JSONErrorBody
	if err := json.Unmarshal(raw, &errBody); err != nil {
		t.Fatal(err)
	}
	if errBody.Error != "cycle already terminal" {
		t.Fatalf("error=%q", errBody.Error)
	}
}

// TestHTTP_postTaskCyclePhase_400ErrorStrings pins phase-start validation.
func TestHTTP_patchTaskCyclePhase_400_path_validation(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskID := handlertest.MustCreateTaskForCycles(t, srv.URL)
	cycleID := mustStartCycle(t, srv.URL, taskID)
	base := srv.URL + "/tasks/" + taskID + "/cycles/" + cycleID + "/phases/"

	cases := []struct {
		name string
		seg  string
		want string
	}{
		{"zero", "0", "phase_seq must be a positive integer"},
		{"negative", "-1", "phase_seq must be a positive integer"},
		{"nonNumeric", "abc", "phase_seq must be a positive integer"},
		{"tooLong", strings.Repeat("9", 33), "phase_seq too long"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, raw := handlertest.DoCyclesRequest(t, http.MethodPatch, base+tc.seg, `{"status":"succeeded"}`)
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status %d body=%s", res.StatusCode, raw)
			}
			var errBody handlertest.JSONErrorBody
			if err := json.Unmarshal(raw, &errBody); err != nil {
				t.Fatal(err)
			}
			if errBody.Error != tc.want {
				t.Fatalf("error=%q want %q", errBody.Error, tc.want)
			}
		})
	}
}

// TestHTTP_patchTaskCyclePhase_400ErrorStrings pins phase-complete status
// validation including the body-level enum guards.
func TestHTTP_patchTaskCyclePhase_400ErrorStrings(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskID := handlertest.MustCreateTaskForCycles(t, srv.URL)
	cycleID := mustStartCycle(t, srv.URL, taskID)
	_, phRaw := handlertest.DoCyclesRequest(t, http.MethodPost,
		srv.URL+"/tasks/"+taskID+"/cycles/"+cycleID+"/phases", `{"phase":"execute"}`)
	var ph taskCyclePhaseResponse
	if err := json.Unmarshal(phRaw, &ph); err != nil {
		t.Fatal(err)
	}
	url := srv.URL + "/tasks/" + taskID + "/cycles/" + cycleID + "/phases/" + strItoa(ph.PhaseSeq)

	cases := []struct {
		name string
		body string
		want string
	}{
		{"unknownField", `{"status":"succeeded","oops":1}`, `json: unknown field "oops"`},
		{"emptyBody", `{}`, "status must be a terminal phase status"},
		{"runningStatus", `{"status":"running"}`, "status must be a terminal phase status"},
		{"invalidEnum", `{"status":"nope"}`, "status must be a terminal phase status"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, raw := handlertest.DoCyclesRequest(t, http.MethodPatch, url, tc.body)
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status %d body=%s", res.StatusCode, raw)
			}
			var errBody handlertest.JSONErrorBody
			if err := json.Unmarshal(raw, &errBody); err != nil {
				t.Fatal(err)
			}
			if errBody.Error != tc.want {
				t.Fatalf("error=%q want %q", errBody.Error, tc.want)
			}
		})
	}
}

// TestHTTP_cycle_path_segment_caps pins the 128-byte abuse guard for {id}
// and {cycleId}. The same cap is documented for every other task route.
