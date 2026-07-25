package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/orchestration"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// applyExecuteEffects persists execute phase outcome and optional terminal
// cycle/task transitions. Store ordering: CompletePhase before TerminateCycle.
func (h *Harness) applyExecuteEffects(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	execPhase *cyclesdomain.TaskCyclePhase,
	result runner.Result,
	effects orchestration.ExecuteEffects,
	commitCount int,
	snap git.PhaseSnapshot,
	operatorCancelled bool,
) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.applyExecuteEffects",
		"task_id", task.ID, "cycle_id", cycle.ID, "continue", effects.ContinueToVerify,
		"terminate", effects.TerminateFailed, "stop", effects.StopLoop)
	if effects.StopLoop {
		h.handleShutdownAfterRun(state, task.ID)
		return false
	}

	result = overlayOperatorCancelOnResult(result, operatorCancelled, effects)
	if effects.TerminateFailed && effects.ResultSummary != "" {
		result.Summary = effects.ResultSummary
	}

	phaseStatus := executePhaseStatusFromEffects(effects)
	phaseDetails := mergeRunnerDetailsWithGit(detailsBytes(result), snap, commitCount)
	phaseDetails = git.MergeCriteriaReportProbeErr(phaseDetails, state.verify.reportParseErr)

	if effects.ContinueToVerify && state.verify.verifySnap.Enabled {
		corr := cyclesdomain.RunCorrelationIDFromDetailsJSON(execPhase.DetailsJSON)
		ev := runner.SetupProgressEvent(runner.ProgressRunStateHandoffVerify, "Handing off to verify…")
		h.persistProgress(parentCtx, task.ID, cycle.ID, execPhase.PhaseSeq, ev)
		h.publishProgress(task.ID, cycle.ID, execPhase.PhaseSeq, corr, ev)
	}

	if !h.completeExecutePhase(parentCtx, state, cycle, execPhase, phaseStatus, result, phaseDetails) {
		h.bestEffortTerminate(parentCtx, state, task.ID, cyclesdomain.CycleStatusFailed, completePhaseFailedReason)
		return false
	}

	if effects.ContinueToVerify {
		return true
	}

	if !effects.TerminateFailed {
		return false
	}

	if !h.terminateCycle(parentCtx, state, cycle.TaskID, cyclesdomain.CycleStatusFailed, string(effects.Reason)) {
		return false
	}
	if effects.TransitionTask == taskcoredomain.StatusFailed {
		if !h.transitionTask(parentCtx, task.ID, effects.TransitionTask, "final_task_transition") {
			return false
		}
	}
	return false
}

func (h *Harness) applyVerifyEffects(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	effects orchestration.VerifyEffects,
	terminalReason string,
) (retryLoop, terminalFailure bool) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.applyVerifyEffects",
		"task_id", task.ID, "cycle_id", cycle.ID)
	if effects.Tampered {
		if !h.terminateCycle(parentCtx, state, cycle.TaskID, cyclesdomain.CycleStatusFailed, verifyTamperedReason) {
			return false, true
		}
		if !h.transitionTask(parentCtx, task.ID, taskcoredomain.StatusFailed, "final_task_transition") {
			return false, true
		}
		return false, true
	}
	if effects.RetryLoop {
		state.verify.verifyAttempt++
		return true, false
	}
	if effects.TerminalFailure {
		if !h.terminateCycle(parentCtx, state, cycle.TaskID, cyclesdomain.CycleStatusFailed, terminalReason) {
			return false, true
		}
		if !h.transitionTask(parentCtx, task.ID, taskcoredomain.StatusFailed, "final_task_transition") {
			return false, true
		}
		return false, true
	}
	return false, false
}

func (h *Harness) applyVerifyDisabledLegacyEffects(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	effects orchestration.VerifyEffects,
) (retryLoop, terminalFailure bool) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.applyVerifyDisabledLegacyEffects",
		"task_id", task.ID, "cycle_id", cycle.ID)
	if !effects.TerminalFailure {
		return false, false
	}
	return h.applyVerifyEffects(parentCtx, task, cycle, state, effects, checklistCompletionFailedReason)
}

func (h *Harness) applyFinalizeEffects(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	effects orchestration.FinalizeEffects,
) bool {
	if !h.terminateCycle(parentCtx, state, cycle.TaskID, effects.CycleStatus, string(effects.Reason)) {
		return false
	}
	if !h.transitionTask(parentCtx, task.ID, effects.TaskStatus, "final_task_transition") {
		return false
	}
	if effects.TaskStatus == taskcoredomain.StatusDone {
		h.emitOnTaskDone(parentCtx, task, cycle.ID)
	}
	if effects.CycleStatus != cyclesdomain.CycleStatusSucceeded {
		return true
	}
	h.publish(task.ID, cycle.ID)
	slog.Info("agent harness run complete", "cmd", calltrace.LogCmd,
		"operation", "agent.harness.Harness.runCycleLoop.summary",
		"task_id", task.ID, "cycle_id", cycle.ID, "attempt_seq", cycle.AttemptSeq,
		"terminal_cycle_status", string(effects.CycleStatus), "task_status", string(effects.TaskStatus),
		"runner", h.runner.Name(), "runner_version", h.runner.Version(),
		"duration_ms", h.opts.Clock().Sub(state.cycle.startedAt).Milliseconds())
	return true
}
