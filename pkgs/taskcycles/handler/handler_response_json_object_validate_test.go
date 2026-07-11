package handler

import (
	"encoding/json"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"testing"
	"time"
)

// docs/api.md pins the response invariant for cycle / phase / event
// JSON columns: meta, details, and data are ALWAYS a JSON object (never
// null, never a string/number/array/bool, never the literal "null"). The
// store-side chokepoint (normalizeJSONObject) was tightened in earlier
// sessions, but legacy rows from before that chokepoint can still carry
// non-object literals in meta_json / details_json / data_json. The response
// builder is the last line of defense before bytes hit the wire — if it
// only normalizes len()==0, those legacy rows leak verbatim to clients,
// which then crash on `Object.entries(meta)` (TypeError on null/scalar)
// or render garbage.
//
// These tests pin "non-object literals are normalized to {} on the way out"
// for every response builder that exposes a JSON-object column.

// nonObjectJSONFixtures enumerates every wire-shape that violates the
// "always a JSON object" invariant. Listed inline (not table-driven) so a
// new violation later (eg. an empty string `""`) is added explicitly.
func nonObjectJSONFixtures() map[string][]byte {
	return map[string][]byte{
		"nil":                nil,
		"empty":              []byte{},
		"whitespace":         []byte("   \n\t"),
		"json_null":          []byte("null"),
		"padded_null":        []byte("  null  "),
		"json_string":        []byte(`"hi"`),
		"json_number":        []byte("123"),
		"json_array":         []byte(`[1,2,3]`),
		"json_bool_true":     []byte("true"),
		"json_bool_false":    []byte("false"),
		"malformed_unclosed": []byte(`{"k":`),
	}
}

func assertObjectMessage(t *testing.T, label string, raw json.RawMessage) {
	t.Helper()
	if len(raw) == 0 {
		t.Fatalf("%s: empty json.RawMessage (want a JSON object literal like {})", label)
	}
	if string(raw) == "null" {
		t.Fatalf("%s: emitted JSON null (docs/api.md cycle/phase/event invariant: always a JSON object)", label)
	}
	var probe any
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatalf("%s: emitted invalid JSON %q: %v (docs invariant: always a valid JSON object)", label, string(raw), err)
	}
	if _, ok := probe.(map[string]any); !ok {
		t.Fatalf("%s: emitted non-object JSON %q (docs invariant: always a JSON object)", label, string(raw))
	}
}

func TestTaskCycleResponseFromDomain_normalizes_non_object_meta(t *testing.T) {
	for name, raw := range nonObjectJSONFixtures() {
		t.Run(name, func(t *testing.T) {
			c := &cyclesdomain.TaskCycle{
				ID:          "cyc_1",
				TaskID:      "tsk_1",
				AttemptSeq:  1,
				Status:      cyclesdomain.CycleStatusRunning,
				StartedAt:   time.Now().UTC(),
				TriggeredBy: string(taskcoredomain.ActorUser),
				MetaJSON:    raw,
			}
			resp := taskCycleResponseFromDomain(c)
			assertObjectMessage(t, "taskCycleResponse.Meta", resp.Meta)
		})
	}
}

func TestTaskCyclePhaseResponseFromDomain_normalizes_non_object_details(t *testing.T) {
	for name, raw := range nonObjectJSONFixtures() {
		t.Run(name, func(t *testing.T) {
			p := &cyclesdomain.TaskCyclePhase{
				ID:          "phs_1",
				CycleID:     "cyc_1",
				Phase:       cyclesdomain.PhaseExecute,
				PhaseSeq:    1,
				Status:      cyclesdomain.PhaseStatusRunning,
				StartedAt:   time.Now().UTC(),
				DetailsJSON: raw,
			}
			resp := taskCyclePhaseResponseFromDomain(p)
			assertObjectMessage(t, "taskCyclePhaseResponse.Details", resp.Details)
		})
	}
}

func TestTaskCycleDetailFromDomain_normalizes_non_object_meta_and_phase_details(t *testing.T) {
	for name, raw := range nonObjectJSONFixtures() {
		t.Run(name, func(t *testing.T) {
			c := &cyclesdomain.TaskCycle{
				ID:          "cyc_2",
				TaskID:      "tsk_2",
				AttemptSeq:  1,
				Status:      cyclesdomain.CycleStatusRunning,
				StartedAt:   time.Now().UTC(),
				TriggeredBy: string(taskcoredomain.ActorUser),
				MetaJSON:    raw,
			}
			phases := []cyclesdomain.TaskCyclePhase{{
				ID:          "phs_2",
				CycleID:     "cyc_2",
				Phase:       cyclesdomain.PhaseExecute,
				PhaseSeq:    1,
				Status:      cyclesdomain.PhaseStatusRunning,
				StartedAt:   time.Now().UTC(),
				DetailsJSON: raw,
			}}
			resp := taskCycleDetailFromDomain(c, phases)
			assertObjectMessage(t, "taskCycleDetailResponse.Meta", resp.Meta)
			if len(resp.Phases) != 1 {
				t.Fatalf("expected 1 phase, got %d", len(resp.Phases))
			}
			assertObjectMessage(t, "taskCycleDetailResponse.Phases[0].Details", resp.Phases[0].Details)
		})
	}
}

func TestTaskCycleResponseFromDomain_object_passes_through(t *testing.T) {
	c := &cyclesdomain.TaskCycle{
		ID:          "cyc_ok",
		TaskID:      "tsk_ok",
		AttemptSeq:  1,
		Status:      cyclesdomain.CycleStatusRunning,
		StartedAt:   time.Now().UTC(),
		TriggeredBy: string(taskcoredomain.ActorUser),
		MetaJSON:    []byte(`{"runner":"cursor-cli","prompt_hash":"abc"}`),
	}
	resp := taskCycleResponseFromDomain(c)
	assertObjectMessage(t, "taskCycleResponse.Meta", resp.Meta)
	var got map[string]any
	if err := json.Unmarshal(resp.Meta, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["runner"] != "cursor-cli" {
		t.Fatalf("runner=%v want cursor-cli (object pass-through must preserve fields)", got["runner"])
	}
	if got["prompt_hash"] != "abc" {
		t.Fatalf("prompt_hash=%v want abc (object pass-through must preserve fields)", got["prompt_hash"])
	}
}
