package verify

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"encoding/json"
	"fmt"
	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	"log/slog"
	"strings"
)

type verifyCriterionPayload struct {
	CriterionID  string `json:"criterion_id"`
	Text         string `json:"text,omitempty"`
	Verified     bool   `json:"verified"`
	VerifierKind string `json:"verifier_kind,omitempty"`
	Reasoning    string `json:"reasoning,omitempty"`
	Evidence     string `json:"evidence,omitempty"`
}

type verifySnapshotPayload struct {
	AttemptSeq  int64                    `json:"attempt_seq"`
	PassedCount int                      `json:"passed_count"`
	FailedCount int                      `json:"failed_count"`
	Criteria    []verifyCriterionPayload `json:"criteria"`
}

type verifyPhaseDetailsPayload struct {
	Verification     verifySnapshotPayload `json:"verification"`
	MirrorDegraded   *bool                 `json:"mirror_degraded,omitempty"`
	VerifyRetryCount *int                  `json:"verify_retry_count,omitempty"`
}

// PhaseDetailsOpts carries optional verify phase metadata persisted in details_json.
type PhaseDetailsOpts struct {
	MirrorDegraded   bool
	VerifyRetryCount int
}

func criterionTextIndex(items []checklistcontract.ChecklistVerifyItem) map[string]string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.criterionTextIndex", "items", len(items))
	out := make(map[string]string, len(items))
	for _, it := range items {
		out[it.ID] = it.Text
	}
	return out
}

func countVerdictOutcome(verdicts []Verdict) (passed, failed int) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.countVerdictOutcome", "verdicts", len(verdicts))
	for _, v := range verdicts {
		if v.Passed {
			passed++
		} else {
			failed++
		}
	}
	return passed, failed
}

// FormatPhaseSummary builds human-readable verify phase.summary for audit mirrors.
func FormatPhaseSummary(
	criteria []checklistcontract.ChecklistVerifyItem,
	verdicts []Verdict,
	succeeded bool,
) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.FormatPhaseSummary",
		"criteria", len(criteria), "verdicts", len(verdicts), "succeeded", succeeded)
	textByID := criterionTextIndex(criteria)
	n := len(verdicts)
	if n == 0 {
		if succeeded {
			return "verify complete"
		}
		return "verification failed"
	}
	passed, failed := countVerdictOutcome(verdicts)
	if succeeded {
		return fmt.Sprintf("All %d criteria verified", passed)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d of %d criteria failed", failed, n)
	for _, v := range verdicts {
		if v.Passed {
			continue
		}
		text := textByID[v.ID]
		if text == "" {
			text = v.ID
		}
		b.WriteString("\n\n- ")
		b.WriteString(text)
		if v.Reasoning != "" {
			b.WriteString(" — ")
			b.WriteString(v.Reasoning)
		}
	}
	return b.String()
}

// EncodePhaseDetails returns structured verify phase details JSON for phase rows.
func EncodePhaseDetails(
	attemptSeq int64,
	criteria []checklistcontract.ChecklistVerifyItem,
	verdicts []Verdict,
	opts PhaseDetailsOpts,
) []byte {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.EncodePhaseDetails",
		"attempt_seq", attemptSeq, "criteria", len(criteria), "verdicts", len(verdicts),
		"mirror_degraded", opts.MirrorDegraded, "verify_retry_count", opts.VerifyRetryCount)
	textByID := criterionTextIndex(criteria)
	passed, failed := countVerdictOutcome(verdicts)
	rows := make([]verifyCriterionPayload, 0, len(verdicts))
	for _, v := range verdicts {
		row := verifyCriterionPayload{
			CriterionID: v.ID,
			Text:        textByID[v.ID],
			Verified:    v.Passed,
		}
		if v.Verifier != "" {
			row.VerifierKind = string(v.Verifier)
		}
		if v.Reasoning != "" {
			row.Reasoning = v.Reasoning
		}
		if v.Evidence != "" {
			row.Evidence = v.Evidence
		}
		rows = append(rows, row)
	}
	payload := verifyPhaseDetailsPayload{
		Verification: verifySnapshotPayload{
			AttemptSeq:  attemptSeq,
			PassedCount: passed,
			FailedCount: failed,
			Criteria:    rows,
		},
	}
	if opts.MirrorDegraded {
		v := true
		payload.MirrorDegraded = &v
	}
	retry := opts.VerifyRetryCount
	payload.VerifyRetryCount = &retry
	b, err := json.Marshal(payload)
	if err != nil {
		return []byte("{}")
	}
	return b
}

// ParseVerifyRetryCount reads verify_retry_count from a verify phase details_json
// payload. The second return is false when the field is absent (callers fall back
// to attempt_seq from verify report rows).
//
//funclogmeasure:skip category=hot-path reason="Pure JSON parser; verify phase persistence traces at store chokepoints."
func ParseVerifyRetryCount(detailsJSON []byte) (int, bool) {
	if len(detailsJSON) == 0 {
		return 0, false
	}
	var payload verifyPhaseDetailsPayload
	if err := json.Unmarshal(detailsJSON, &payload); err != nil {
		return 0, false
	}
	if payload.VerifyRetryCount == nil {
		return 0, false
	}
	return *payload.VerifyRetryCount, true
}

// ParseMirrorDegraded reports whether mirror_degraded was set in verify phase details.
//
//funclogmeasure:skip category=hot-path reason="Pure JSON parser; verify phase persistence traces at store chokepoints."
func ParseMirrorDegraded(detailsJSON []byte) bool {
	if len(detailsJSON) == 0 {
		return false
	}
	var payload verifyPhaseDetailsPayload
	if err := json.Unmarshal(detailsJSON, &payload); err != nil {
		return false
	}
	return payload.MirrorDegraded != nil && *payload.MirrorDegraded
}
