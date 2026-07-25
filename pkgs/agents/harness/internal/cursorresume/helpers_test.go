package cursorresume

import (
	"testing"

	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestSessionPhaseForResume(t *testing.T) {
	t.Parallel()
	if got := SessionPhaseForResume(cyclesdomain.PhaseVerify); got != cyclesdomain.PhaseExecute {
		t.Fatalf("verify → %q, want execute", got)
	}
	if got := SessionPhaseForResume(cyclesdomain.PhaseExecute); got != cyclesdomain.PhaseExecute {
		t.Fatalf("execute → %q, want execute", got)
	}
}
