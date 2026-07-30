package verify

import (
	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
)

const failedReasonPrefix = "verification_failed"

// FailedReasonPrefix is the stable terminate_reason prefix for verification failures.
const FailedReasonPrefix = failedReasonPrefix

// Snapshot captures checklist criteria for one task run.
type Snapshot struct {
	Enabled  bool
	Criteria []checklistcontract.ChecklistVerifyItem
}

// Verdict is the harness-internal outcome for one criterion after claim acceptance.
type Verdict struct {
	ID        string
	Passed    bool
	Evidence  string
	Verifier  checklistdomain.VerifierKind
	Reasoning string
}

// TamperedError is retained for resume/error typing compatibility. Claim-only
// acceptance no longer runs verify-phase integrity checks.
type TamperedError struct {
	Reason string
}

func (e *TamperedError) Error() string {
	if e == nil {
		return ""
	}
	return "verify_tampered: " + e.Reason
}
