package orchestration

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/verify"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
)

// RetryMode selects whether the next loop iteration re-runs execute.
type RetryMode string

const (
	RetryModeVerifyOnly    RetryMode = "verify_only"
	RetryModeFullReexecute RetryMode = "full_reexecute"
)

// ReasonCode is a stable eligibility reason for logs and tests (ADR-0028).
type ReasonCode string

const (
	ReasonVerifyOnlyInfra              ReasonCode = "verify_only_infra"
	ReasonFullReexecuteImplementation  ReasonCode = "full_reexecute_implementation"
	ReasonFullReexecuteReportInvalid   ReasonCode = "full_reexecute_report_invalid"
	ReasonFullReexecuteHeadChanged     ReasonCode = "full_reexecute_head_changed"
	ReasonFullReexecuteIngestFailed    ReasonCode = "full_reexecute_ingest_failed"
	ReasonFullReexecuteExecuteNotReady ReasonCode = "full_reexecute_execute_not_reached"
)

// VerifyRetryFailureClass groups verify failures for retry policy (ADR-0028).
// Distinct from resume.ContinuationFailureKind (continuation prompt taxonomy).
type VerifyRetryFailureClass int

const (
	VerifyRetryFailureInfra VerifyRetryFailureClass = iota
	VerifyRetryFailureImplementation
)

// ClassifyInput carries pre-fetched booleans for execute-validity gates.
type ClassifyInput struct {
	FailureClass         VerifyRetryFailureClass
	CriteriaReportValid  bool
	GitHeadMatchesAnchor bool
	CommitIngestOK       bool
	ExecuteReachedVerify bool
}

// ClassifyFailureClass maps verify outcomes to infra vs implementation.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ClassifyFailureClass(verdicts []verify.Verdict, pipelineFailed bool) VerifyRetryFailureClass {
	for _, v := range verdicts {
		if v.Passed {
			continue
		}
		switch v.Verifier {
		case checklistdomain.VerifierAgentSelf, checklistdomain.VerifierVerifyAgent:
			return VerifyRetryFailureImplementation
		}
	}
	if pipelineFailed {
		return VerifyRetryFailureInfra
	}
	return VerifyRetryFailureImplementation
}

// ClassifyVerifyRetryMode implements the ADR-0028 decision table.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ClassifyVerifyRetryMode(in ClassifyInput) (RetryMode, ReasonCode) {
	if !in.ExecuteReachedVerify {
		return RetryModeFullReexecute, ReasonFullReexecuteExecuteNotReady
	}
	if !in.CriteriaReportValid {
		return RetryModeFullReexecute, ReasonFullReexecuteReportInvalid
	}
	if !in.GitHeadMatchesAnchor {
		return RetryModeFullReexecute, ReasonFullReexecuteHeadChanged
	}
	if !in.CommitIngestOK {
		return RetryModeFullReexecute, ReasonFullReexecuteIngestFailed
	}
	if in.FailureClass == VerifyRetryFailureImplementation {
		return RetryModeFullReexecute, ReasonFullReexecuteImplementation
	}
	return RetryModeVerifyOnly, ReasonVerifyOnlyInfra
}
