package cycles_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
)

func TestHTTP_getTaskCycles_400ErrorStrings(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskID := handlertest.MustCreateTaskForCycles(t, srv.URL)

	cases := []struct {
		name  string
		query string
		want  string
	}{
		{"limitOver200", "?limit=999", "limit must be integer 0..200"},
		{"limitNegative", "?limit=-1", "limit must be integer 0..200"},
		{"limitNonNumeric", "?limit=nope", "limit must be integer 0..200"},
		{"limitTooLong", "?limit=" + strings.Repeat("9", 33), "limit too long"},
		{"beforeZero", "?before_attempt_seq=0", "before_attempt_seq must be a positive integer"},
		{"beforeNegative", "?before_attempt_seq=-1", "before_attempt_seq must be a positive integer"},
		{"beforeNonNumeric", "?before_attempt_seq=nope", "before_attempt_seq must be a positive integer"},
		{"beforeTooLong", "?before_attempt_seq=" + strings.Repeat("9", 33), "before_attempt_seq too long"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, raw := handlertest.DoCyclesRequest(t, http.MethodGet, srv.URL+"/tasks/"+taskID+"/cycles"+tc.query, "")
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status %d (want 400) body=%s", res.StatusCode, raw)
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

// TestHTTP_getTaskCycles_response_shape pins the envelope for the empty
// case (no cycles yet). Cycles must be `[]`, never null or omitted.
func TestHTTP_getTaskCycles_response_shape(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskID := handlertest.MustCreateTaskForCycles(t, srv.URL)

	res, raw := handlertest.DoCyclesRequest(t, http.MethodGet, srv.URL+"/tasks/"+taskID+"/cycles", "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d body=%s", res.StatusCode, raw)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"task_id", "cycles", "limit", "has_more"} {
		if _, ok := doc[k]; !ok {
			t.Fatalf("missing key %q (full=%v)", k, doc)
		}
	}
	if string(doc["cycles"]) != "[]" {
		t.Fatalf("empty cycles must serialise as [], got %s", doc["cycles"])
	}
	if string(doc["has_more"]) != "false" {
		t.Fatalf("has_more must be false on empty list, got %s", doc["has_more"])
	}
	if _, ok := doc["next_before_attempt_seq"]; ok {
		t.Fatalf("next_before_attempt_seq must be omitted when has_more=false (omitempty contract); got %s", doc["next_before_attempt_seq"])
	}
}

// TestHTTP_getTaskCycles_keysetCursor pins the keyset pagination wire
// contract added in Session 30: ?before_attempt_seq= MUST return cycles
// strictly older than the cursor (never the cursor row itself, which
// would duplicate across pages), and the response envelope MUST include
// next_before_attempt_seq when has_more=true so clients can paginate
// without re-implementing the cursor calculation themselves. Mirrors
// the store-side keyset test (TestStore_ListCyclesForTaskBefore_keysetCursor)
// at the HTTP layer so a future handler refactor that drops the cursor
// query param parse, the limit+1 fan-out, or the cursor write-back
// would fail loudly.
func TestHTTP_getTaskCycles_keysetCursor(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskID := handlertest.MustCreateTaskForCycles(t, srv.URL)
	listURL := srv.URL + "/tasks/" + taskID + "/cycles"
	// At-most-one-running means we have to terminate each cycle before
	// starting the next; seed three completed cycles so the cursor test has
	// real attempt_seq values to walk.
	for i := 0; i < 3; i++ {
		cid := mustStartCycle(t, srv.URL, taskID)
		res, raw := handlertest.DoCyclesRequest(t, http.MethodPatch,
			srv.URL+"/tasks/"+taskID+"/cycles/"+cid,
			`{"status":"succeeded"}`)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("seed terminate cycle #%d: %d body=%s", i+1, res.StatusCode, raw)
		}
	}

	t.Run("firstPageHasNextCursor", func(t *testing.T) {
		res, raw := handlertest.DoCyclesRequest(t, http.MethodGet, listURL+"?limit=1", "")
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status %d body=%s", res.StatusCode, raw)
		}
		var env taskCyclesListResponse
		if err := json.Unmarshal(raw, &env); err != nil {
			t.Fatal(err)
		}
		if len(env.Cycles) != 1 {
			t.Fatalf("first page len = %d, want 1", len(env.Cycles))
		}
		if !env.HasMore {
			t.Fatalf("has_more must be true when more pages exist")
		}
		if env.NextBeforeAttemptSeq == nil || *env.NextBeforeAttemptSeq != env.Cycles[0].AttemptSeq {
			t.Fatalf("next_before_attempt_seq = %v, want pointer to %d", env.NextBeforeAttemptSeq, env.Cycles[0].AttemptSeq)
		}
	})

	t.Run("strictlyLessThanCursor", func(t *testing.T) {
		res, raw := handlertest.DoCyclesRequest(t, http.MethodGet, listURL+"?limit=1", "")
		if res.StatusCode != http.StatusOK {
			t.Fatalf("seed: status %d", res.StatusCode)
		}
		var first taskCyclesListResponse
		if err := json.Unmarshal(raw, &first); err != nil {
			t.Fatal(err)
		}
		if first.NextBeforeAttemptSeq == nil {
			t.Fatalf("seed expected next_before_attempt_seq")
		}
		nextURL := listURL + "?limit=200&before_attempt_seq=" + strconv.FormatInt(*first.NextBeforeAttemptSeq, 10)
		res, raw = handlertest.DoCyclesRequest(t, http.MethodGet, nextURL, "")
		if res.StatusCode != http.StatusOK {
			t.Fatalf("next page: status %d body=%s", res.StatusCode, raw)
		}
		var next taskCyclesListResponse
		if err := json.Unmarshal(raw, &next); err != nil {
			t.Fatal(err)
		}
		for _, c := range next.Cycles {
			if c.AttemptSeq >= *first.NextBeforeAttemptSeq {
				t.Fatalf("cursor row leaked: attempt_seq=%d should be < %d (strict <)", c.AttemptSeq, *first.NextBeforeAttemptSeq)
			}
		}
		if next.HasMore {
			t.Fatalf("has_more must be false on last page (only 2 cycles past cursor in a 3-cycle seed); got envelope=%+v", next)
		}
		if next.NextBeforeAttemptSeq != nil {
			t.Fatalf("next_before_attempt_seq must be omitted on last page; got %d", *next.NextBeforeAttemptSeq)
		}
	})
}

// TestHTTP_getTaskCycle_404_when_taskMismatch protects against cross-task
// id smuggling: a valid cycle id under a different task must surface as
// 404, not the foreign cycle's contents.
func TestHTTP_getTaskCycle_404_when_taskMismatch(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskA := handlertest.MustCreateTaskForCycles(t, srv.URL)
	taskB := handlertest.MustCreateTaskForCycles(t, srv.URL)
	cycleID := mustStartCycle(t, srv.URL, taskA)

	res, raw := handlertest.DoCyclesRequest(t, http.MethodGet, srv.URL+"/tasks/"+taskB+"/cycles/"+cycleID, "")
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("cross-task GET cycle: status %d body=%s", res.StatusCode, raw)
	}
}

// TestHTTP_patchTaskCycle_404_when_taskMismatch covers the same protection
// for the terminate path: a foreign cycle id must not be mutated.
func TestHTTP_getTaskCycle_phase_response_shape(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	taskID := handlertest.MustCreateTaskForCycles(t, srv.URL)
	cycleID := mustStartCycle(t, srv.URL, taskID)
	handlertest.DoCyclesRequest(t, http.MethodPost,
		srv.URL+"/tasks/"+taskID+"/cycles/"+cycleID+"/phases", `{"phase":"execute"}`)

	res, raw := handlertest.DoCyclesRequest(t, http.MethodGet, srv.URL+"/tasks/"+taskID+"/cycles/"+cycleID, "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d body=%s", res.StatusCode, raw)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"id", "task_id", "attempt_seq", "status", "started_at", "triggered_by", "meta", "phases"} {
		if _, ok := doc[k]; !ok {
			t.Fatalf("missing top-level key %q (full=%v)", k, doc)
		}
	}
	var phases []map[string]json.RawMessage
	if err := json.Unmarshal(doc["phases"], &phases); err != nil {
		t.Fatalf("decode phases: %v raw=%s", err, doc["phases"])
	}
	if len(phases) != 1 {
		t.Fatalf("phases=%d want 1", len(phases))
	}
	for _, k := range []string{"id", "cycle_id", "phase", "phase_seq", "status", "started_at", "details", "event_seq"} {
		if _, ok := phases[0][k]; !ok {
			t.Fatalf("phases[0] missing key %q (full=%v)", k, phases[0])
		}
	}
	for _, k := range []string{"ended_at", "summary"} {
		if _, ok := phases[0][k]; ok {
			t.Fatalf("phases[0] optional key %q must be omitted on running phase (full=%v)", k, phases[0])
		}
	}
}

// SSE publishing for cycle and phase mutations is asserted by the
// `task_cycle_changed` subtests in TestHTTP_SSE_triggerSurface
// (sse_trigger_surface_test.go). The Stage 4 "must emit zero SSE" guardrail
// was retired in Stage 5 once the trigger surface was lit up.
