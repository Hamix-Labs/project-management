package domain

import (
	"encoding/json"
	"strings"
)

// PhaseDetailsRunCorrelationID is the details_json key for a per-phase log
// correlation handle (ADR-0030). One id is minted at StartPhase and preserved
// through CompletePhase.
const PhaseDetailsRunCorrelationID = "run_correlation_id"

// PhaseDetailsSessionID is the details_json key for the Cursor CLI session_id
// captured from stream-json (ADR-0031). The harness writes this key at two
// moments: (1) once, best-effort, when the cursor adapter first observes a
// non-empty session_id on the live NDJSON stream (Layer B early patch); and
// (2) on any terminal phase outcome — success, failure, timeout, cancel, exec
// error — when the adapter extracts the id from captured stdout (Layer A).
// PatchPhaseDetails enforces first-wins for the mid-run patch so the id is
// stable across resume decisions even if a later frame reports a different id.
const PhaseDetailsSessionID = "session_id"

// RunCorrelationIDFromDetailsJSON extracts run_correlation_id from phase
// details_json. Returns "" when absent or malformed.
//
//funclogmeasure:skip category=hot-path reason="Pure JSON extract without I/O."
func RunCorrelationIDFromDetailsJSON(detailsJSON []byte) string {
	if len(detailsJSON) == 0 {
		return ""
	}
	var obj map[string]any
	if err := json.Unmarshal(detailsJSON, &obj); err != nil {
		return ""
	}
	id, _ := obj[PhaseDetailsRunCorrelationID].(string)
	return id
}

// SessionIDFromDetailsJSON extracts session_id from phase details_json.
// Returns "" when absent or malformed.
//
//funclogmeasure:skip category=hot-path reason="Pure JSON extract without I/O."
func SessionIDFromDetailsJSON(detailsJSON []byte) string {
	if len(detailsJSON) == 0 {
		return ""
	}
	var obj map[string]any
	if err := json.Unmarshal(detailsJSON, &obj); err != nil {
		return ""
	}
	id, _ := obj[PhaseDetailsSessionID].(string)
	return strings.TrimSpace(id)
}
