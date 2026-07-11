package harness

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestSelectRecoveryKind_criteriaReportInvalidFromState(t *testing.T) {
	t.Parallel()
	h := &Harness{}
	state := &processState{
		verify: verifyLifecycleState{reportParseErr: "criteria report invalid: unknown field function"},
	}
	kind := h.selectRecoveryKind(cyclesdomain.PhaseExecute, state, cycleLoopOpts{}, taskcoredomain.RetryFresh)
	if kind != prompt.RecoveryCriteriaReportInvalid {
		t.Fatalf("kind=%q want %q", kind, prompt.RecoveryCriteriaReportInvalid)
	}
}

func TestSelectRecoveryKind_operatorRetryDefersToCriteriaProbeErr(t *testing.T) {
	t.Parallel()
	h := &Harness{}
	state := &processState{
		verify: verifyLifecycleState{reportParseErr: "criteria report missing"},
	}
	opts := cycleLoopOpts{continuation: &ContinuationBundle{ParentCycleID: "parent-1"}}
	kind := h.selectRecoveryKind(cyclesdomain.PhaseExecute, state, opts, taskcoredomain.RetryResume)
	if kind != prompt.RecoveryCriteriaReportMissing {
		t.Fatalf("kind=%q want criteria_report_missing", kind)
	}
}
