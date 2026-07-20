package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"errors"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

type startCycleOpts struct {
	parentCycleID *string
	retryMode     taskcoredomain.RetryMode
}

// RunWithRetry starts a new cycle. intent==nil is the existing first-run path.
func (h *Harness) RunWithRetry(parentCtx context.Context, task *taskcoredomain.Task, intent *taskcoredomain.PendingRetry) {
	if intent == nil {
		h.runFreshCycle(parentCtx, task, startCycleOpts{})
		return
	}
	if err := intent.Validate(); err != nil {
		slog.Warn("agent harness retry intent invalid", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.RunWithRetry.invalid_intent",
			"task_id", task.ID, "err", err)
		h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_invalid_intent")
		return
	}
	switch intent.Mode {
	case taskcoredomain.RetryFresh:
		h.runFreshRetry(parentCtx, task, intent)
	case taskcoredomain.RetryResume:
		h.runResumeRetry(parentCtx, task, intent)
	default:
		h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_invalid_intent")
	}
}

func (h *Harness) runFreshRetry(parentCtx context.Context, task *taskcoredomain.Task, intent *taskcoredomain.PendingRetry) {
	if _, err := h.gitResetForFreshRetry(parentCtx, intent.ParentCycleID); err != nil {
		reason := retryGitResetFailed
		if errors.Is(err, git.ErrRetryResetAnchorMissing) {
			reason = retryResetAnchorMissing
		}
		slog.Warn("agent harness fresh retry git reset failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runFreshRetry.reset_err",
			"task_id", task.ID, "parent_cycle_id", intent.ParentCycleID, "err", err)
		h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, reason)
		return
	}
	parentID := intent.ParentCycleID
	h.runFreshCycle(parentCtx, task, startCycleOpts{
		parentCycleID: &parentID,
		retryMode:     taskcoredomain.RetryFresh,
	})
}

func (h *Harness) runResumeRetry(parentCtx context.Context, task *taskcoredomain.Task, intent *taskcoredomain.PendingRetry) {
	cp, err := h.loadCheckpointFromParent(parentCtx, intent.ParentCycleID)
	if err != nil {
		slog.Warn("agent harness resume retry checkpoint failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runResumeRetry.checkpoint_err",
			"task_id", task.ID, "parent_cycle_id", intent.ParentCycleID, "err", err)
		h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_checkpoint_failed")
		return
	}
	startedAt := h.opts.Clock()
	state := processState{
		cycle: cycleLifecycleState{startedAt: startedAt},
		verify: verifyLifecycleState{
			previouslyPassed: harnessVerdictsFromResume(cp.PreviouslyPassed),
			verifyAttempt:    0,
			verifyFeedback:   cp.VerifyFeedback,
		},
	}
	defer h.recoverFromPanic(&state, *task)

	parentID := intent.ParentCycleID
	cycle, ok := h.startCycle(parentCtx, task, &state, startCycleOpts{
		parentCycleID: &parentID,
		retryMode:     taskcoredomain.RetryResume,
	})
	if !ok {
		h.bestEffortFailTask(parentCtx, task.ID)
		return
	}
	if cp.Entry == resumeEntryVerifyOnly {
		if err := h.resumeSvc().SeedCrossCycleExecuteFromParent(parentCtx, cycle, intent.ParentCycleID); err != nil {
			slog.Warn("agent harness verify-only resume seed execute failed", "cmd", calltrace.LogCmd,
				"operation", "agent.harness.Harness.runResumeRetry.seed_execute_err",
				"task_id", task.ID, "parent_cycle_id", intent.ParentCycleID, "err", err)
			h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_verify_only_seed_failed")
			return
		}
		if err := h.resumeSvc().MirrorParentCriteriaForVerifyOnly(parentCtx, cycle.ID, intent.ParentCycleID); err != nil {
			slog.Warn("agent harness verify-only resume mirror criteria failed", "cmd", calltrace.LogCmd,
				"operation", "agent.harness.Harness.runResumeRetry.mirror_criteria_err",
				"task_id", task.ID, "parent_cycle_id", intent.ParentCycleID, "err", err)
			h.resumeSvc().FailTaskAfterRetryPrep(parentCtx, task.ID, "retry_verify_only_mirror_failed")
			return
		}
	}
	snap, err := h.loadVerificationSnapshot(parentCtx, task.ID)
	if err != nil {
		slog.Error("agent harness resume-retry verification snapshot failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runResumeRetry.verify_snap_err",
			"task_id", task.ID, "cycle_id", cycle.ID, "err", err)
		h.bestEffortTerminate(parentCtx, &state, task.ID, cyclesdomain.CycleStatusFailed, "verification_snapshot_load_failed")
		return
	}
	state.verify.verifySnap = snap
	h.runCycleLoop(parentCtx, task, cycle, &state, cycleLoopOpts{
		resumeNotice:     true,
		interruptedPhase: cyclesdomain.PhaseExecute,
		skipFirstExecute: cp.Entry == resumeEntryVerifyOnly,
		knownCommits:     cp.KnownCommits,
		continuation:     cp.Continuation,
	})
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) runFreshCycle(parentCtx context.Context, task *taskcoredomain.Task, opts startCycleOpts) {
	startedAt := h.opts.Clock()
	state := processState{
		cycle:  cycleLifecycleState{startedAt: startedAt},
		verify: verifyLifecycleState{previouslyPassed: map[string]criterionVerdict{}},
	}
	defer h.recoverFromPanic(&state, *task)

	cycle, ok := h.startCycle(parentCtx, task, &state, opts)
	if !ok {
		h.bestEffortFailTask(parentCtx, task.ID)
		return
	}
	snap, err := h.loadVerificationSnapshot(parentCtx, task.ID)
	if err != nil {
		slog.Error("agent harness fresh-cycle verification snapshot failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runFreshCycle.verify_snap_err",
			"task_id", task.ID, "cycle_id", cycle.ID, "err", err)
		h.bestEffortTerminate(parentCtx, &state, task.ID, cyclesdomain.CycleStatusFailed, "verification_snapshot_load_failed")
		return
	}
	state.verify.verifySnap = snap
	h.runCycleLoop(parentCtx, task, cycle, &state, cycleLoopOpts{})
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) reconstructCheckpoint(ctx context.Context, cycle *cyclesdomain.TaskCycle) (resumeCheckpoint, error) {
	return h.resumeSvc().ReconstructCheckpoint(ctx, cycle)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) loadCheckpointFromParent(ctx context.Context, parentCycleID string) (resumeCheckpoint, error) {
	return h.resumeSvc().LoadCheckpointFromParent(ctx, parentCycleID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) loadContinuationBundle(ctx context.Context, parentCycleID string) (ContinuationBundle, error) {
	return h.resumeSvc().LoadContinuationBundle(ctx, parentCycleID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) seedCrossCycleExecuteFromParent(ctx context.Context, cycle *cyclesdomain.TaskCycle, parentCycleID string) error {
	return h.resumeSvc().SeedCrossCycleExecuteFromParent(ctx, cycle, parentCycleID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) mirrorParentCriteriaForVerifyOnly(ctx context.Context, childCycleID, parentCycleID string) error {
	return h.resumeSvc().MirrorParentCriteriaForVerifyOnly(ctx, childCycleID, parentCycleID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) failTaskAfterRetryPrep(ctx context.Context, taskID, reason string) {
	h.resumeSvc().FailTaskAfterRetryPrep(ctx, taskID, reason)
}
