package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/cursorresume"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/orchestration"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/verify"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

type cycleLoopOpts struct {
	resumeNotice     bool
	interruptedPhase cyclesdomain.Phase
	skipFirstExecute bool
	knownCommits     []cyclesdomain.TaskCycleCommit
	continuation     *ContinuationBundle
	skipVerify       bool
}

func (h *Harness) composeExecutePrompt(ctx context.Context, task *taskcoredomain.Task, cycle *cyclesdomain.TaskCycle, state *processState, opts cycleLoopOpts) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.composeExecutePrompt",
		"task_id", task.ID, "cycle_id", cycle.ID, "resume_notice", opts.resumeNotice)
	promptText := task.InitialPrompt
	promptText = prompt.InjectCriteria(
		promptText,
		checklistItemsForPrompt(state.verify.verifySnap.Criteria),
		reports.CriteriaReportPath(h.opts.ReportDir, cycle.ID),
		lockedPassIDSet(state.verify.lockedPasses),
		h.agentMCPActive(ctx),
	)
	retryMode := retryModeFromCycleMeta(cycle)
	runKind := runKindFromCycleMeta(cycle)
	if runKind == taskcoredomain.PendingKindPolish {
		promptText = prompt.AppendPolishNotice(promptText, cycle, polishNoticeInputFromCycle(cycle, state, opts.knownCommits))
	} else if bundle := opts.continuation; bundle != nil {
		promptText = prompt.ComposeContinuation(promptText, continuationInputFromBundle(cycle, bundle))
		if bundle.ExecuteFeedback != "" {
			promptText = prompt.AppendExecuteHarnessFeedback(promptText, bundle.ExecuteFeedback)
		}
	} else if opts.resumeNotice {
		if retryMode == taskcoredomain.RetryResume {
			promptText = prompt.AppendOperatorRetryResumeNotice(promptText, cycle, opts.knownCommits)
		} else {
			promptText = prompt.AppendResumeNotice(promptText, cycle, opts.interruptedPhase, opts.knownCommits)
		}
	}
	if !state.git.gitSnap.Skipped {
		operatorResume := retryMode == taskcoredomain.RetryResume || runKind == taskcoredomain.PendingKindPolish
		promptText = prompt.AppendGitCommitPolicy(promptText, operatorResume)
	}
	return promptText
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func polishNoticeInputFromCycle(cycle *cyclesdomain.TaskCycle, state *processState, known []cyclesdomain.TaskCycleCommit) prompt.PolishNoticeInput {
	textByID := make(map[string]string, len(state.verify.verifySnap.Criteria))
	for _, c := range state.verify.verifySnap.Criteria {
		textByID[c.ID] = c.Text
	}
	flaggedIDs := polishFlaggedIDsFromCycleMeta(cycle)
	newIDs := polishNewIDsFromCycleMeta(cycle)
	flagged := make([]prompt.PolishCriterion, 0, len(flaggedIDs))
	for _, id := range flaggedIDs {
		flagged = append(flagged, prompt.PolishCriterion{ID: id, Text: textByID[id]})
	}
	newRows := make([]prompt.PolishCriterion, 0, len(newIDs))
	for _, id := range newIDs {
		newRows = append(newRows, prompt.PolishCriterion{ID: id, Text: textByID[id]})
	}
	return prompt.PolishNoticeInput{
		Instructions: polishInstructionsFromCycleMeta(cycle),
		SkipVerify:   polishSkipVerifyFromCycleMeta(cycle),
		Flagged:      flagged,
		New:          newRows,
		KnownCommits: known,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func recordLockedPassVerdicts(state *processState, verdicts []criterionVerdict) {
	for _, v := range verdicts {
		if !v.Passed {
			continue
		}
		if _, exists := state.verify.lockedPasses[v.ID]; !exists {
			state.verify.lockedPasses[v.ID] = v
		}
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func unionLockedPassesVerdicts(state *processState) []criterionVerdict {
	unionVerdicts := make([]criterionVerdict, 0, len(state.verify.lockedPasses))
	for _, v := range state.verify.lockedPasses {
		unionVerdicts = append(unionVerdicts, v)
	}
	return unionVerdicts
}

// runCycleLoopExecute runs one execute phase iteration. Returns false when
// runCycleLoop should return immediately.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) runCycleLoopExecute(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	opts cycleLoopOpts,
) bool {
	phaseOut := h.executeSvc().RunPhase(parentCtx, task, cycle, h.executePhasePorts(state, opts))
	if phaseOut.FatalReason != "" {
		h.bestEffortTerminate(parentCtx, state, task.ID, cyclesdomain.CycleStatusFailed, phaseOut.FatalReason)
		return false
	}

	snap := phaseOut.Snap
	state.git.gitSnap = snap
	execPhase := phaseOut.ExecPhase
	result := phaseOut.Result
	operatorCancelled := phaseOut.OperatorCancelled

	effects := orchestration.DecideExecutePostRun(phaseOut.PostRunInput)
	result, effects = h.enforceExecuteSessionID(parentCtx, result, effects, state)
	h.probeCriteriaReport(state, cycle.ID)
	cont := h.applyExecuteEffects(parentCtx, task, cycle, state, execPhase, result, effects, phaseOut.CommitCount, snap, operatorCancelled)
	if cont {
		h.anchorPostExecuteState(parentCtx, state, execPhase.PhaseSeq, snap, phaseOut.IngestAttempted, phaseOut.IngestOutcome, phaseOut.IngestErr)
	}
	return cont
}

// enforceExecuteSessionID fails a successful Cursor execute that omitted session_id
// when resume is enabled (same-chat hard requirement).
func (h *Harness) enforceExecuteSessionID(
	ctx context.Context,
	result runner.Result,
	effects orchestration.ExecuteEffects,
	state *processState,
) (runner.Result, orchestration.ExecuteEffects) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.enforceExecuteSessionID",
		"continue_to_claim_acceptance", effects.ContinueToClaimAcceptance)
	if !effects.ContinueToClaimAcceptance {
		return result, effects
	}
	if h.runner == nil || !isCursorSessionRunner(h.runner) {
		return result, effects
	}
	settings, err := h.store.GetSettings(ctx)
	if err != nil || !settings.CursorSessionResumeEnabled {
		return result, effects
	}
	if cyclesdomain.SessionIDFromDetailsJSON(detailsBytes(result)) != "" {
		return result, effects
	}
	hf := cursorresume.MissingSessionAfterExecute()
	result.Details = mergeFailureDetails(detailsBytes(result), hf.DetailsMap())
	if result.Summary == "" {
		result.Summary = "Cursor chat session id missing"
	}
	return result, orchestration.ExecuteEffects{
		TerminateFailed: true,
		TransitionTask:  taskcoredomain.StatusFailed,
		Reason:          orchestration.ReasonCursorMissingSessionID,
		ResultSummary:   hf.Explain(),
	}
}

//funclogmeasure:skip category=hot-path reason="Pure runner name check without I/O."
func isCursorSessionRunner(r runner.Runner) bool {
	if r == nil {
		return false
	}
	n := strings.ToLower(strings.TrimSpace(r.Name()))
	return n == "cursor" || n == "cursor-cli" || strings.HasPrefix(n, "cursor")
}

// runCycleLoopClaimAcceptance runs claim acceptance once per cycle (ADR-0092 one-shot).
// terminalFailure is true when claim acceptance failed terminally (caller should return).
func (h *Harness) runCycleLoopClaimAcceptance(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
) (terminalFailure bool) {
	if !state.verify.verifySnap.Enabled {
		checklistErr := h.completeChecklistLegacy(parentCtx, task.ID)
		if checklistErr != nil {
			slog.Warn("agent harness checklist completion failed",
				"cmd", calltrace.LogCmd,
				"operation", "agent.harness.Harness.runCycleLoop.checklist_err",
				"task_id", task.ID, "err", checklistErr)
		}
		effects := orchestration.DecideVerifyDisabledLegacy(checklistErr)
		_, term := h.applyVerifyEffects(parentCtx, task, cycle, state, effects, checklistCompletionFailedReason)
		return term
	}

	verdicts, verifyErr := h.runVerificationPipeline(parentCtx, task, cycle, state, state.verify.verifySnap)
	if verifyErr == nil {
		recordLockedPassVerdicts(state, verdicts)
		return false
	}

	state.verify.lastFailedVerdicts = append([]criterionVerdict(nil), verdicts...)
	var tampered *verify.TamperedError
	if errors.As(verifyErr, &tampered) {
		state.verify.reportTampered = true
	}

	if hf, ok := cursorresume.AsHardFail(verifyErr); ok {
		effects := orchestration.VerifyEffects{TerminalFailure: true}
		_, term := h.applyVerifyEffects(parentCtx, task, cycle, state, effects, hf.FormatReason())
		return term
	}

	tamperedResult := errors.As(verifyErr, &tampered)
	effects := orchestration.VerifyEffects{TerminalFailure: true, Tampered: tamperedResult}
	if tamperedResult {
		slog.Info("agent harness verify terminal (tampered)", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runCycleLoopClaimAcceptance.one_shot",
			"task_id", task.ID, "cycle_id", cycle.ID)
	} else {
		slog.Info("agent harness verify terminal (one-shot)", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runCycleLoopClaimAcceptance.one_shot",
			"task_id", task.ID, "cycle_id", cycle.ID)
	}
	terminalReason := formatVerificationFailedReason(verdicts, state.verify.lockedPasses)
	_, term := h.applyVerifyEffects(parentCtx, task, cycle, state, effects, terminalReason)
	return term
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin wrapper; runCycleLoopFinalizeSuccessOpts emits finalize traces."
func (h *Harness) runCycleLoopFinalizeSuccess(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
) {
	h.runCycleLoopFinalizeSuccessOpts(parentCtx, task, cycle, state, true)
}

func (h *Harness) runCycleLoopFinalizeSuccessOpts(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	applyCompletions bool,
) {
	var completionErr error
	if applyCompletions {
		unionVerdicts := unionLockedPassesVerdicts(state)
		completionErr = h.applyVerifiedCompletions(parentCtx, task.ID, cycle.ID, unionVerdicts)
	}
	effects := orchestration.DecideFinalizeSuccess(completionErr)
	if completionErr != nil {
		slog.Warn("agent harness checklist completion failed",
			"cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runCycleLoop.finalize_err",
			"task_id", task.ID, "err", completionErr)
	}
	if ok := h.applyFinalizeEffects(parentCtx, task, cycle, state, effects); !ok {
		slog.Error("agent harness finalize effects incomplete", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runCycleLoopFinalizeSuccess.effects_incomplete",
			"task_id", task.ID, "cycle_id", cycle.ID,
			"cycle_status", string(effects.CycleStatus), "task_status", string(effects.TaskStatus))
		if state.cycle.cycleStarted {
			h.bestEffortTerminate(parentCtx, state, task.ID, cyclesdomain.CycleStatusFailed, "finalize_effects_failed")
		} else {
			_ = h.transitionTask(parentCtx, task.ID, effects.TaskStatus, "finalize_effects_retry")
		}
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) runCycleLoop(parentCtx context.Context, task *taskcoredomain.Task, cycle *cyclesdomain.TaskCycle, state *processState, opts cycleLoopOpts) {
	state.resume.continuation = opts.continuation
	state.resume.resumeNotice = opts.resumeNotice
	state.resume.interruptedPhase = opts.interruptedPhase
	if bundle := opts.continuation; bundle != nil {
		state.verify.reportParseErr = strings.TrimSpace(bundle.CriteriaReportProbeErr)
	}
	if !opts.skipFirstExecute {
		if !h.runCycleLoopExecute(parentCtx, task, cycle, state, opts) {
			return
		}
	}
	if opts.skipVerify {
		h.runCycleLoopFinalizeSuccessOpts(parentCtx, task, cycle, state, false)
		return
	}
	if h.runCycleLoopClaimAcceptance(parentCtx, task, cycle, state) {
		return
	}
	h.runCycleLoopFinalizeSuccess(parentCtx, task, cycle, state)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) repoRootForGit(_ context.Context) string {
	return strings.TrimSpace(h.opts.WorkingDir)
}
