package orchestration

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/verify"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
)

func TestClassify_EC01_verifyInfra_verifyOnly(t *testing.T) {
	t.Parallel()
	mode, reason := ClassifyVerifyRetryMode(ClassifyInput{
		FailureClass:         VerifyRetryFailureInfra,
		CriteriaReportValid:  true,
		GitHeadMatchesAnchor: true,
		CommitIngestOK:       true,
		ExecuteReachedVerify: true,
	})
	if mode != RetryModeVerifyOnly || reason != ReasonVerifyOnlyInfra {
		t.Fatalf("got mode=%q reason=%q", mode, reason)
	}
}

func TestClassify_EC02_verifyAgentReject_fullReexecute(t *testing.T) {
	t.Parallel()
	cls := ClassifyFailureClass([]verify.Verdict{{Passed: false, Verifier: checklistdomain.VerifierVerifyAgent}}, false)
	if cls != VerifyRetryFailureImplementation {
		t.Fatalf("class=%v", cls)
	}
	mode, reason := ClassifyVerifyRetryMode(ClassifyInput{
		FailureClass:         cls,
		CriteriaReportValid:  true,
		GitHeadMatchesAnchor: true,
		CommitIngestOK:       true,
		ExecuteReachedVerify: true,
	})
	if mode != RetryModeFullReexecute || reason != ReasonFullReexecuteImplementation {
		t.Fatalf("got mode=%q reason=%q", mode, reason)
	}
}

func TestClassify_EC03_claimedNotDone_fullReexecute(t *testing.T) {
	t.Parallel()
	cls := ClassifyFailureClass([]verify.Verdict{{Passed: false, Verifier: checklistdomain.VerifierAgentSelf}}, false)
	mode, _ := ClassifyVerifyRetryMode(ClassifyInput{
		FailureClass:         cls,
		CriteriaReportValid:  true,
		GitHeadMatchesAnchor: true,
		CommitIngestOK:       true,
		ExecuteReachedVerify: true,
	})
	if mode != RetryModeFullReexecute {
		t.Fatalf("mode=%q", mode)
	}
}

func TestClassify_EC04_reportInvalid_fullReexecute(t *testing.T) {
	t.Parallel()
	mode, reason := ClassifyVerifyRetryMode(ClassifyInput{
		FailureClass:         VerifyRetryFailureInfra,
		CriteriaReportValid:  false,
		GitHeadMatchesAnchor: true,
		CommitIngestOK:       true,
		ExecuteReachedVerify: true,
	})
	if mode != RetryModeFullReexecute || reason != ReasonFullReexecuteReportInvalid {
		t.Fatalf("got mode=%q reason=%q", mode, reason)
	}
}

func TestClassify_EC05_headChanged_fullReexecute(t *testing.T) {
	t.Parallel()
	mode, reason := ClassifyVerifyRetryMode(ClassifyInput{
		FailureClass:         VerifyRetryFailureInfra,
		CriteriaReportValid:  true,
		GitHeadMatchesAnchor: false,
		CommitIngestOK:       true,
		ExecuteReachedVerify: true,
	})
	if mode != RetryModeFullReexecute || reason != ReasonFullReexecuteHeadChanged {
		t.Fatalf("got mode=%q reason=%q", mode, reason)
	}
}

func TestClassify_EC06_ingestFailed_fullReexecute(t *testing.T) {
	t.Parallel()
	mode, reason := ClassifyVerifyRetryMode(ClassifyInput{
		FailureClass:         VerifyRetryFailureInfra,
		CriteriaReportValid:  true,
		GitHeadMatchesAnchor: true,
		CommitIngestOK:       false,
		ExecuteReachedVerify: true,
	})
	if mode != RetryModeFullReexecute || reason != ReasonFullReexecuteIngestFailed {
		t.Fatalf("got mode=%q reason=%q", mode, reason)
	}
}

func TestClassify_EC08_budgetExhausted_terminal(t *testing.T) {
	t.Parallel()
	e := DecideVerifyRetryWithValidity(3, 3, VerifyResultFailRetryable, true)
	if !e.TerminalFailure || e.RetryLoop || e.SkipNextExecute {
		t.Fatalf("effects=%+v", e)
	}
}

func TestClassifyFailureClass_pipelineErrIsInfra(t *testing.T) {
	t.Parallel()
	if ClassifyFailureClass(nil, true) != VerifyRetryFailureInfra {
		t.Fatal("want infra for pipeline failure without verdicts")
	}
}
