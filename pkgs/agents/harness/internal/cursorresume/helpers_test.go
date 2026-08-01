package cursorresume

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestSelectRecoveryKind_openPR(t *testing.T) {
	t.Parallel()
	got := SelectRecoveryKind(RecoveryKindInput{
		Phase:   cyclesdomain.PhaseExecute,
		RunKind: taskcoredomain.PendingKindOpenPR,
	})
	if got != prompt.RecoveryHumanOpenPR {
		t.Fatalf("got %q want %q", got, prompt.RecoveryHumanOpenPR)
	}
}

func TestSessionPhaseForResume(t *testing.T) {
	t.Parallel()
	if got := SessionPhaseForResume(cyclesdomain.PhaseVerify); got != cyclesdomain.PhaseExecute {
		t.Fatalf("PhaseVerify: got %q want execute", got)
	}
	if got := SessionPhaseForResume(cyclesdomain.PhaseExecute); got != cyclesdomain.PhaseExecute {
		t.Fatalf("PhaseExecute: got %q want execute", got)
	}
}
