package harness

import (
	"testing"

	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestCycleLoopOptsFromCheckpoint_interrupt(t *testing.T) {
	t.Parallel()
	commits := []cyclesdomain.TaskCycleCommit{{SHA: "abc"}}

	opts := cycleLoopOptsFromCheckpoint(resumeCheckpoint{
		Entry:        resumeEntryExecute,
		KnownCommits: commits,
	}, cycleLoopEntryInterrupt)
	if !opts.resumeNotice || opts.skipFirstExecute || opts.interruptedPhase != cyclesdomain.PhaseExecute {
		t.Fatalf("execute interrupt opts: %+v", opts)
	}

	opts = cycleLoopOptsFromCheckpoint(resumeCheckpoint{
		Entry: resumeEntryVerifyOnly,
	}, cycleLoopEntryInterrupt)
	if opts.resumeNotice || !opts.skipFirstExecute || opts.interruptedPhase != cyclesdomain.PhaseVerify {
		t.Fatalf("verify-only interrupt opts: %+v", opts)
	}

	opts = cycleLoopOptsFromCheckpoint(resumeCheckpoint{
		Entry: resumeEntryAfterExecuteSuccess,
	}, cycleLoopEntryInterrupt)
	if opts.resumeNotice || !opts.skipFirstExecute || opts.interruptedPhase != "" {
		t.Fatalf("after-execute interrupt opts: %+v", opts)
	}
}

func TestCycleLoopOptsFromCheckpoint_operatorRetry(t *testing.T) {
	t.Parallel()
	bundle := &ContinuationBundle{ParentCycleID: "parent"}
	opts := cycleLoopOptsFromCheckpoint(resumeCheckpoint{
		Entry:        resumeEntryVerifyOnly,
		Continuation: bundle,
	}, cycleLoopEntryOperatorRetry)
	if !opts.resumeNotice || !opts.skipFirstExecute || opts.interruptedPhase != cyclesdomain.PhaseExecute {
		t.Fatalf("verify-only operator retry opts: %+v", opts)
	}
	if opts.continuation != bundle {
		t.Fatal("expected continuation bundle retained")
	}

	opts = cycleLoopOptsFromCheckpoint(resumeCheckpoint{
		Entry: resumeEntryExecute,
	}, cycleLoopEntryOperatorRetry)
	if !opts.resumeNotice || opts.skipFirstExecute {
		t.Fatalf("execute operator retry opts: %+v", opts)
	}
}

func TestVerifyLifecycleFromCheckpoint_resetAttempt(t *testing.T) {
	t.Parallel()
	cp := resumeCheckpoint{
		VerifyAttempt:  3,
		VerifyFeedback: "fix",
		PreviouslyPassed: map[string]criterionVerdict{
			"c1": {ID: "c1", Passed: true},
		},
	}
	got := verifyLifecycleFromCheckpoint(cp, true)
	if got.verifyAttempt != 0 || got.verifyFeedback != "fix" || !got.previouslyPassed["c1"].Passed {
		t.Fatalf("reset attempt: %+v", got)
	}
	got = verifyLifecycleFromCheckpoint(cp, false)
	if got.verifyAttempt != 3 {
		t.Fatalf("preserve attempt: %+v", got)
	}
}
