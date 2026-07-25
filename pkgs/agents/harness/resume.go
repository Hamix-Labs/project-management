package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"log/slog"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
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
		h.handleResumeCheckpointFailure(parentCtx, task.ID, cycle.ID)
		return
	}

	state := processState{
		cycle: cycleLifecycleState{
			cycleID:        cycle.ID,
			cycleStarted:   true,
			startedAt:      startedAt,
			effectiveModel: effectiveModelFromCycleMeta(h.runner, task, cycle),
		},
		verify: verifyLifecycleFromCheckpoint(cp, false),
	}
	defer h.recoverFromPanic(&state, *task)

	slog.Info("agent harness resume branch", "cmd", calltrace.LogCmd,
		"operation", "agent.harness.Harness.Resume.branch",
		"task_id", task.ID, "cycle_id", cycle.ID,
		"entry", cp.Entry, "locked_count", len(cp.PreviouslyPassed),
		"verify_attempt", cp.VerifyAttempt)

	h.enterCycleLoopFromCheckpoint(parentCtx, task, cycle, &state, cp, cycleLoopEntryInterrupt)
}

// handleResumeCheckpointFailure avoids clobbering a successful finalize when a
// late duplicate Resume sees an already-succeeded verify phase / terminal cycle.
func (h *Harness) handleResumeCheckpointFailure(ctx context.Context, taskID, cycleID string) {
	if cycleID != "" {
		fresh, err := h.store.GetCycle(ctx, cycleID)
		if err == nil && fresh != nil {
			if fresh.Status == cyclesdomain.CycleStatusSucceeded {
				slog.Info("agent harness resume checkpoint failed after succeeded cycle; healing",
					"cmd", calltrace.LogCmd,
					"operation", "agent.harness.Harness.Resume.checkpoint_err.succeeded_cycle",
					"task_id", taskID, "cycle_id", cycleID)
				h.healTaskReviewAfterSucceededCycle(ctx, taskID, cycleID)
				return
			}
			if fresh.Status == cyclesdomain.CycleStatusRunning && h.verifyPhaseAlreadySucceeded(ctx, cycleID) {
				slog.Info("agent harness resume checkpoint failed; verify already succeeded, leaving owner to finalize",
					"cmd", calltrace.LogCmd,
					"operation", "agent.harness.Harness.Resume.checkpoint_err.verify_already_succeeded",
					"task_id", taskID, "cycle_id", cycleID)
				return
			}
		}
		if err != nil && !errors.Is(err, taskcoredomain.ErrNotFound) {
			slog.Warn("agent harness resume GetCycle failed", "cmd", calltrace.LogCmd,
				"operation", "agent.harness.Harness.Resume.checkpoint_err.get_cycle",
				"task_id", taskID, "cycle_id", cycleID, "err", err)
		}
	}
	cur, err := h.store.Get(ctx, taskID)
	if err == nil && cur != nil {
		switch cur.Status {
		case taskcoredomain.StatusReview, taskcoredomain.StatusDone:
			slog.Info("agent harness resume checkpoint failed; task already terminal success",
				"cmd", calltrace.LogCmd,
				"operation", "agent.harness.Harness.Resume.checkpoint_err.already_terminal",
				"task_id", taskID, "status", string(cur.Status))
			return
		}
	} else if err != nil && !errors.Is(err, taskcoredomain.ErrNotFound) {
		slog.Warn("agent harness resume Get task failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.Resume.checkpoint_err.get_task",
			"task_id", taskID, "err", err)
	}
	h.bestEffortFailTask(ctx, taskID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) verifyPhaseAlreadySucceeded(ctx context.Context, cycleID string) bool {
	phases, err := h.store.ListPhasesForCycle(ctx, cycleID)
	if err != nil || len(phases) == 0 {
		return false
	}
	last := phases[len(phases)-1]
	return last.Phase == cyclesdomain.PhaseVerify && last.Status == cyclesdomain.PhaseStatusSucceeded
}

func (h *Harness) healTaskReviewAfterSucceededCycle(ctx context.Context, taskID, cycleID string) {
	cur, err := h.store.Get(ctx, taskID)
	if err != nil {
		if !errors.Is(err, taskcoredomain.ErrNotFound) {
			slog.Warn("agent harness heal Get failed", "cmd", calltrace.LogCmd,
				"operation", "agent.harness.Harness.healTaskReviewAfterSucceededCycle.get_err",
				"task_id", taskID, "err", err)
		}
		return
	}
	if cur.Status == taskcoredomain.StatusReview || cur.Status == taskcoredomain.StatusDone {
		return
	}
	if cur.Status != taskcoredomain.StatusRunning && cur.Status != taskcoredomain.StatusFailed {
		return
	}
	review := taskcoredomain.StatusReview
	if _, err := h.store.Update(ctx, taskID, taskcorestore.UpdateTaskInput{Status: &review}, taskcoredomain.ActorAgent); err != nil {
		if !errors.Is(err, taskcoredomain.ErrNotFound) {
			slog.Warn("agent harness heal running/failed→review failed", "cmd", calltrace.LogCmd,
				"operation", "agent.harness.Harness.healTaskReviewAfterSucceededCycle.update_err",
				"task_id", taskID, "cycle_id", cycleID, "err", err)
		}
		return
	}
	h.publishTaskUpdated(taskID)
	slog.Info("agent harness healed task to review after succeeded cycle",
		"cmd", calltrace.LogCmd,
		"operation", "agent.harness.Harness.healTaskReviewAfterSucceededCycle",
		"task_id", taskID, "cycle_id", cycleID, "prev_status", string(cur.Status))
}
