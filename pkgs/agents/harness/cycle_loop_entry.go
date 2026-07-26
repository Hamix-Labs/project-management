package harness

import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// cycleLoopEntryKind selects how a checkpoint seeds cycleLoopOpts.
type cycleLoopEntryKind int

const (
	// cycleLoopEntryInterrupt is worker crash/interrupt resume of the same cycle.
	cycleLoopEntryInterrupt cycleLoopEntryKind = iota
	// cycleLoopEntryOperatorRetry is operator resume-retry into a new child cycle.
	cycleLoopEntryOperatorRetry
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func verifyLifecycleFromCheckpoint(cp resumeCheckpoint, resetVerifyAttempt bool) verifyLifecycleState {
	attempt := cp.VerifyAttempt
	if resetVerifyAttempt {
		attempt = 0
	}
	return verifyLifecycleState{
		previouslyPassed: cloneVerdictMap(cp.PreviouslyPassed),
		verifyAttempt:    attempt,
		verifyFeedback:   cp.VerifyFeedback,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func cycleLoopOptsFromCheckpoint(cp resumeCheckpoint, kind cycleLoopEntryKind) cycleLoopOpts {
	opts := cycleLoopOpts{knownCommits: cp.KnownCommits}
	switch kind {
	case cycleLoopEntryInterrupt:
		switch cp.Entry {
		case resumeEntryExecute:
			opts.resumeNotice = true
			opts.interruptedPhase = cyclesdomain.PhaseExecute
		case resumeEntryVerifyOnly:
			opts.skipFirstExecute = true
			opts.interruptedPhase = cyclesdomain.PhaseVerify
		case resumeEntryAfterExecuteSuccess:
			opts.skipFirstExecute = true
		}
	case cycleLoopEntryOperatorRetry:
		opts.resumeNotice = true
		opts.interruptedPhase = cyclesdomain.PhaseExecute
		opts.skipFirstExecute = cp.Entry == resumeEntryVerifyOnly
		opts.continuation = cp.Continuation
	}
	return opts
}

// enterCycleLoopFromCheckpoint loads the verify snapshot and enters runCycleLoop
// with opts derived from the checkpoint entry kind.
func (h *Harness) enterCycleLoopFromCheckpoint(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	cp resumeCheckpoint,
	kind cycleLoopEntryKind,
) {
	snap, err := h.loadVerificationSnapshot(parentCtx, task)
	if err != nil {
		slog.Error("agent harness verification snapshot failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.enterCycleLoopFromCheckpoint.verify_snap_err",
			"task_id", task.ID, "cycle_id", cycle.ID, "err", err)
		h.bestEffortTerminate(parentCtx, state, task.ID, cyclesdomain.CycleStatusFailed, "verification_snapshot_load_failed")
		return
	}
	state.verify.verifySnap = snap
	h.runCycleLoop(parentCtx, task, cycle, state, cycleLoopOptsFromCheckpoint(cp, kind))
}
