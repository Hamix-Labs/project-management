package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"log/slog"
)

// Resume continues an in-flight cycle after process interruption. The task
// must already be StatusRunning and cycle must be StatusRunning. The worker
// calls this after FinalizeInterruptedPhases and queue admission.
func (h *Harness) Resume(parentCtx context.Context, task *taskcoredomain.Task, cycle *cyclesdomain.TaskCycle) {
	slog.Info("agent harness resume", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.Resume",
		"task_id", task.ID, "cycle_id", cycle.ID, "attempt_seq", cycle.AttemptSeq)
	startedAt := h.opts.Clock()
	cp, err := h.reconstructCheckpoint(parentCtx, cycle)
	if err != nil {
		slog.Warn("agent harness resume checkpoint failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.Resume.checkpoint_err",
			"task_id", task.ID, "cycle_id", cycle.ID, "err", err)
		h.bestEffortFailTask(parentCtx, task.ID)
		return
	}

	state := processState{
		cycle: cycleLifecycleState{
			cycleID:        cycle.ID,
			cycleStarted:   true,
			startedAt:      startedAt,
			effectiveModel: effectiveModelFromCycleMeta(h.runner, task, cycle),
		},
		verify: verifyLifecycleState{
			previouslyPassed: cloneVerdictMap(cp.PreviouslyPassed),
			verifyAttempt:    cp.VerifyAttempt,
			verifyFeedback:   cp.VerifyFeedback,
		},
	}
	snap, err := h.loadVerificationSnapshot(parentCtx, task.ID)
	if err != nil {
		slog.Error("agent harness resume verification snapshot failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.Resume.verify_snap_err",
			"task_id", task.ID, "cycle_id", cycle.ID, "err", err)
		h.bestEffortTerminate(parentCtx, &state, task.ID, cyclesdomain.CycleStatusFailed, "verification_snapshot_load_failed")
		return
	}
	state.verify.verifySnap = snap

	defer h.recoverFromPanic(&state, *task)

	opts := cycleLoopOpts{knownCommits: cp.KnownCommits}
	switch cp.Entry {
	case resumeEntryExecute:
		opts.resumeNotice = true
		opts.interruptedPhase = cyclesdomain.PhaseExecute
	case resumeEntryVerifyOnly:
		opts.resumeNotice = false
		opts.skipFirstExecute = true
		opts.interruptedPhase = cyclesdomain.PhaseVerify
	case resumeEntryAfterExecuteSuccess:
		opts.skipFirstExecute = true
	}

	slog.Info("agent harness resume branch", "cmd", calltrace.LogCmd,
		"operation", "agent.harness.Harness.Resume.branch",
		"task_id", task.ID, "cycle_id", cycle.ID,
		"entry", cp.Entry, "locked_count", len(cp.PreviouslyPassed),
		"verify_attempt", cp.VerifyAttempt)

	h.runCycleLoop(parentCtx, task, cycle, &state, opts)
}
