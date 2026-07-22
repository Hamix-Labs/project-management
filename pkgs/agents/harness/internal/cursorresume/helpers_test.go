package cursorresume

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestFirstVerifyAfterNewExecute(t *testing.T) {
	t.Parallel()
	if !FirstVerifyAfterNewExecute(1, 3) {
		t.Fatal("expected fresh verify after new execute")
	}
	if FirstVerifyAfterNewExecute(3, 3) {
		t.Fatal("expected resume verify within same execute stint")
	}
}

func TestSelectRecoveryKind_criteriaReportInvalid(t *testing.T) {
	t.Parallel()
	kind := SelectRecoveryKind(RecoveryKindInput{
		Phase:          cyclesdomain.PhaseExecute,
		ReportParseErr: "criteria report invalid: unknown field function",
		RetryMode:      taskcoredomain.RetryFresh,
	})
	if kind != prompt.RecoveryCriteriaReportInvalid {
		t.Fatalf("kind=%q want %q", kind, prompt.RecoveryCriteriaReportInvalid)
	}
}

func TestSelectRecoveryKind_operatorRetryDefersToCriteriaProbeErr(t *testing.T) {
	t.Parallel()
	kind := SelectRecoveryKind(RecoveryKindInput{
		Phase:           cyclesdomain.PhaseExecute,
		ReportParseErr:  "criteria report missing",
		RetryMode:       taskcoredomain.RetryResume,
		HasContinuation: true,
	})
	if kind != prompt.RecoveryCriteriaReportMissing {
		t.Fatalf("kind=%q want criteria_report_missing", kind)
	}
}
