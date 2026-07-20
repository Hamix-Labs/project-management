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
