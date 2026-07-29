package verify

import (
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
)

const failedReasonPrefix = "verification_failed"

// FailedReasonPrefix is the stable terminate_reason prefix for verification failures.
const FailedReasonPrefix = failedReasonPrefix

// Snapshot captures verify settings and checklist criteria for one task run.
type Snapshot struct {
	Enabled    bool
	Criteria   []checklistcontract.ChecklistVerifyItem
	// VerifyModel is the optional settings pin for PhaseVerify (--model).
	// Empty means inherit task.CursorModel / execute runner default.
	VerifyModel string
	// VerifyChatMode is the effective chat policy for this run (ADR-0086).
	VerifyChatMode settingsdomain.VerifyChatMode
}

// Verdict is the harness-internal outcome for one criterion after verify work.
type Verdict struct {
	ID        string
	Passed    bool
	Evidence  string
	Verifier  checklistdomain.VerifierKind
	Reasoning string
}

// TamperedError is returned when post-verify integrity detects unauthorized
// working-tree changes. Terminal for the cycle — callers use errors.As and
// map to verify_tampered terminate reason.
type TamperedError struct {
	Reason string
}

func (e *TamperedError) Error() string {
	if e == nil {
		return ""
	}
	return "verify_tampered: " + e.Reason
}
