package harness

import "testing"

func TestFirstVerifyAfterNewExecute(t *testing.T) {
	t.Parallel()
	h := &Harness{}
	state := &processState{phase: phaseLifecycleState{lastCompletedExecutePhaseSeq: 3, lastVerifyAfterExecuteSeq: 1}}
	if !h.firstVerifyAfterNewExecute(state) {
		t.Fatal("expected fresh verify after new execute")
	}
	state.phase.lastVerifyAfterExecuteSeq = 3
	if h.firstVerifyAfterNewExecute(state) {
		t.Fatal("expected resume verify within same execute stint")
	}
}
