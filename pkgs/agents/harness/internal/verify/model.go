package verify

import (
	"encoding/json"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// EffectiveVerifyModel resolves the Cursor --model for PhaseVerify.
// Non-empty snap.VerifyModel wins; otherwise task.CursorModel; empty means
// adapter DefaultCursorModel.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func EffectiveVerifyModel(task *taskcoredomain.Task, snap Snapshot) string {
	if m := strings.TrimSpace(snap.VerifyModel); m != "" {
		return m
	}
	if task == nil {
		return ""
	}
	return strings.TrimSpace(task.CursorModel)
}

// MergeFailureDetailsIntoPhaseJSON adds failure_kind / standardized_message onto
// an existing verify phase details payload (or builds a minimal object).
//
//funclogmeasure:skip category=hot-path reason="Pure JSON merge without I/O."
func MergeFailureDetailsIntoPhaseJSON(base []byte, failureKind, standardizedMessage string) []byte {
	failureKind = strings.TrimSpace(failureKind)
	standardizedMessage = strings.TrimSpace(standardizedMessage)
	if failureKind == "" && standardizedMessage == "" {
		return base
	}
	var m map[string]any
	if len(base) > 0 {
		_ = json.Unmarshal(base, &m)
	}
	if m == nil {
		m = map[string]any{}
	}
	if failureKind != "" {
		m["failure_kind"] = failureKind
	}
	if standardizedMessage != "" {
		m["standardized_message"] = standardizedMessage
	}
	out, err := json.Marshal(m)
	if err != nil {
		return base
	}
	return out
}
