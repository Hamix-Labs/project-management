package harness

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/orchestration"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// Invariant: execute terminal effects never also continue to verify.
func TestInvariant_executeNeverContinuesAfterTerminalFailure(t *testing.T) {
	t.Parallel()
	terminalInputs := []orchestration.ExecutePostRunInput{
		{RunnerOutcome: orchestration.ExecuteRunnerOutcomeTimeout},
		{RunnerOutcome: orchestration.ExecuteRunnerOutcomeOK, OperatorCancelled: true},
		{
			RunnerOutcome: orchestration.ExecuteRunnerOutcomeOK,
			CommitIngest: orchestration.ExecuteCommitIngestSummary{
				IngestAttempted: true,
				IngestErr:       true,
			},
		},
	}
	for i, in := range terminalInputs {
		e := orchestration.DecideExecutePostRun(in)
		if e.TerminateFailed && e.ContinueToVerify {
			t.Fatalf("input %d: terminal and continue both set: %+v", i, e)
		}
		if e.TerminateFailed && e.StopLoop {
			t.Fatalf("input %d: terminal and stop both set: %+v", i, e)
		}
	}
}

// Invariant: terminal cycle statuses are mutually exclusive with an open running phase
// in the domain model — harness termination paths must land on a terminal status.
func TestInvariant_terminalCycleStatusesAreTerminal(t *testing.T) {
	t.Parallel()
	for _, st := range []cyclesdomain.CycleStatus{
		cyclesdomain.CycleStatusSucceeded,
		cyclesdomain.CycleStatusFailed,
		cyclesdomain.CycleStatusAborted,
	} {
		if !cyclesdomain.TerminalCycleStatus(st) {
			t.Fatalf("%q should be terminal", st)
		}
	}
	if cyclesdomain.TerminalCycleStatus(cyclesdomain.CycleStatusRunning) {
		t.Fatal("running must not be terminal")
	}
}
