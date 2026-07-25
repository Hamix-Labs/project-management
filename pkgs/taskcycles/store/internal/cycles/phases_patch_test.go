package cycles

import (
	"encoding/json"
	"testing"

	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestPatchPhaseDetailsJSON_preservesRunCorrelationID(t *testing.T) {
	t.Parallel()

	existing := []byte(`{"run_correlation_id":"corr-1","other":"a"}`)
	incoming := []byte(`{"session_id":"sess-1","run_correlation_id":"clobber"}`)

	out, err := patchPhaseDetailsJSON(existing, incoming)
	if err != nil {
		t.Fatalf("patchPhaseDetailsJSON: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got[cyclesdomain.PhaseDetailsRunCorrelationID] != "corr-1" {
		t.Fatalf("run_correlation_id must be preserved: got %#v", got[cyclesdomain.PhaseDetailsRunCorrelationID])
	}
	if got[cyclesdomain.PhaseDetailsSessionID] != "sess-1" {
		t.Fatalf("session_id must be merged in: got %#v", got[cyclesdomain.PhaseDetailsSessionID])
	}
	if got["other"] != "a" {
		t.Fatalf("prior fields must survive: got %#v", got["other"])
	}
}

func TestPatchPhaseDetailsJSON_sessionIDFirstWins(t *testing.T) {
	t.Parallel()

	existing := []byte(`{"session_id":"sess-first","run_correlation_id":"corr-1"}`)
	incoming := []byte(`{"session_id":"sess-second"}`)

	out, err := patchPhaseDetailsJSON(existing, incoming)
	if err != nil {
		t.Fatalf("patchPhaseDetailsJSON: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got[cyclesdomain.PhaseDetailsSessionID] != "sess-first" {
		t.Fatalf("existing session_id must win (first-wins): got %#v", got[cyclesdomain.PhaseDetailsSessionID])
	}
}

func TestPatchPhaseDetailsJSON_addsSessionIDWhenAbsent(t *testing.T) {
	t.Parallel()

	existing := []byte(`{"run_correlation_id":"corr-1"}`)
	incoming := []byte(`{"session_id":"sess-new"}`)

	out, err := patchPhaseDetailsJSON(existing, incoming)
	if err != nil {
		t.Fatalf("patchPhaseDetailsJSON: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got[cyclesdomain.PhaseDetailsSessionID] != "sess-new" {
		t.Fatalf("session_id must be added when missing: got %#v", got[cyclesdomain.PhaseDetailsSessionID])
	}
	if got[cyclesdomain.PhaseDetailsRunCorrelationID] != "corr-1" {
		t.Fatalf("run_correlation_id must be preserved: got %#v", got[cyclesdomain.PhaseDetailsRunCorrelationID])
	}
}

func TestPatchPhaseDetailsJSON_keepsExistingWhenPatchOmitsSessionID(t *testing.T) {
	t.Parallel()

	existing := []byte(`{"session_id":"sess-first","run_correlation_id":"corr-1"}`)
	incoming := []byte(`{"other":"x"}`)

	out, err := patchPhaseDetailsJSON(existing, incoming)
	if err != nil {
		t.Fatalf("patchPhaseDetailsJSON: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got[cyclesdomain.PhaseDetailsSessionID] != "sess-first" {
		t.Fatalf("session_id must be preserved when patch omits it: got %#v", got[cyclesdomain.PhaseDetailsSessionID])
	}
	if got["other"] != "x" {
		t.Fatalf("patch fields must merge: got %#v", got["other"])
	}
}
