package cursorresume

import (
	"testing"

	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestSessionPhaseForResume(t *testing.T) {
	t.Parallel()
	if got := SessionPhaseForResume(cyclesdomain.PhaseVerify, settingsdomain.VerifyChatModeSameChat); got != cyclesdomain.PhaseExecute {
		t.Fatalf("same_chat verify → %q, want execute", got)
	}
	if got := SessionPhaseForResume(cyclesdomain.PhaseVerify, settingsdomain.VerifyChatModeDifferentChat); got != cyclesdomain.PhaseVerify {
		t.Fatalf("different_chat verify → %q, want verify", got)
	}
	if got := SessionPhaseForResume(cyclesdomain.PhaseExecute, settingsdomain.VerifyChatModeDifferentChat); got != cyclesdomain.PhaseExecute {
		t.Fatalf("execute → %q, want execute", got)
	}
}

func TestFirstVerifyAfterNewExecute(t *testing.T) {
	t.Parallel()
	if !FirstVerifyAfterNewExecute(0, 1) {
		t.Fatal("expected fresh after new execute")
	}
	if FirstVerifyAfterNewExecute(1, 1) {
		t.Fatal("expected not fresh when seq caught up")
	}
}
