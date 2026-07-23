package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/orchestration"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/verify"
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
		verifiedCriterionIDs(state.verify.previouslyPassed),
	)
	promptText = prompt.AppendVerifyFeedback(promptText, state.verify.verifyFeedback)
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
func recordPassedCriterionVerdicts(state *processState, verdicts []criterionVerdict) {
	for _, v := range verdicts {
		if !v.Passed {
			continue
		}
		if _, exists := state.verify.previouslyPassed[v.ID]; !exists {
			state.verify.previouslyPassed[v.ID] = v
		}
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func unionPreviouslyPassedVerdicts(state *processState) []criterionVerdict {
	unionVerdicts := make([]criterionVerdict, 0, len(state.verify.previouslyPassed))
	for _, v := range state.verify.previouslyPassed {
		unionVerdicts = append(unionVerdicts, v)
	}
	return unionVerdicts
}

// runCycleLoopExecute runs one execute phase iteration. Returns false when
// runCycleLoop should return immediately.
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
	if phaseOut.StaleRecovery && effects.ContinueToVerify {
		recovered := streamIdleRecoveredEvent()
		h.persistProgress(parentCtx, task.ID, cycle.ID, execPhase.PhaseSeq, recovered)
		h.publishProgress(task.ID, cycle.ID, execPhase.PhaseSeq, state.phase.runCorrelationID, recovered)
	}
	h.probeCriteriaReport(state, cycle.ID)
	cont := h.applyExecuteEffects(parentCtx, task, cycle, state, execPhase, result, effects, phaseOut.CommitCount, snap, operatorCancelled, phaseOut.StaleRecovery)
	if cont {
		h.anchorPostExecuteState(parentCtx, state, execPhase.PhaseSeq, snap, phaseOut.IngestAttempted, phaseOut.IngestOutcome, phaseOut.IngestErr)
	}
	return cont
}

// runCycleLoopVerify runs verification for one loop iteration. retryLoop is
// true when the outer loop should continue for another execute↔verify attempt.
// skipNextExecute is true when the next iteration should skip execute (ADR-0028).
// terminalFailure is true when verification failed terminally (caller should return).
func (h *Harness) runCycleLoopVerify(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
) (retryLoop bool, terminalFailure bool, skipNextExecute bool) {
	if !state.verify.verifySnap.Enabled {
		checklistErr := h.completeChecklistLegacy(parentCtx, task.ID)
		if checklistErr != nil {
			slog.Warn("agent harness checklist completion failed",
				"cmd", calltrace.LogCmd,
				"operation", "agent.harness.Harness.runCycleLoop.checklist_err",
				"task_id", task.ID, "err", checklistErr)
		}
		effects := orchestration.DecideVerifyDisabledLegacy(checklistErr)
		retry, term := h.applyVerifyEffects(parentCtx, task, cycle, state, effects, checklistCompletionFailedReason)
		return retry, term, false
	}

	verdicts, feedback, verifyErr := h.runVerificationPipeline(parentCtx, task, cycle, state, state.verify.verifySnap, state.verify.verifyFeedback)
	if verifyErr != nil && feedback != "" {
		state.verify.verifyFeedback = feedback
	}
	recordPassedCriterionVerdicts(state, verdicts)
	if verifyErr != nil {
		state.verify.lastFailedVerdicts = append([]criterionVerdict(nil), verdicts...)
		var tampered *verify.TamperedError
		if errors.As(verifyErr, &tampered) {
			state.verify.reportTampered = true
		}
	}
	if verifyErr == nil {
		return false, false, false
	}

	var result orchestration.VerifyResult
	var tampered *verify.TamperedError
	if errors.As(verifyErr, &tampered) {
		result = orchestration.VerifyResultFailTampered
	} else {
		result = orchestration.VerifyResultFailRetryable
	}

	classifyIn := h.gatherRetryClassifyInput(parentCtx, cycle, state, verdicts, verifyErr)
	retryMode, reasonCode := orchestration.ClassifyVerifyRetryMode(classifyIn)
	executeStillValid := retryMode == orchestration.RetryModeVerifyOnly
	effects := orchestration.DecideVerifyRetryWithValidity(state.verify.verifyAttempt, state.verify.verifySnap.MaxRetries, result, executeStillValid)
	if effects.RetryLoop {
		slog.Info("agent harness verify retry classified", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runCycleLoopVerify.retry_mode",
			"task_id", task.ID, "cycle_id", cycle.ID,
			"retry_mode", string(retryMode), "reason_code", string(reasonCode),
			"skip_next_execute", effects.SkipNextExecute)
	}
	terminalReason := formatVerificationFailedReason(verdicts, state.verify.previouslyPassed)
	retry, term := h.applyVerifyEffects(parentCtx, task, cycle, state, effects, terminalReason)
	return retry, term, effects.SkipNextExecute
}

func (h *Harness) runCycleLoopFinalizeSuccess(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
) {
	unionVerdicts := unionPreviouslyPassedVerdicts(state)
	completionErr := h.applyVerifiedCompletions(parentCtx, task.ID, cycle.ID, unionVerdicts)
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
			// Cycle already terminal: retry the intended task status (done/failed).
			// Do not force StatusFailed after a succeeded cycle — that orphans a
			// successful run as "failed" when only the status write failed.
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
	skipExecute := opts.skipFirstExecute
	for {
		if !skipExecute {
			if !h.runCycleLoopExecute(parentCtx, task, cycle, state, opts) {
				return
			}
		} else {
			skipExecute = false
		}

		if opts.skipVerify {
			h.runCycleLoopFinalizeSuccess(parentCtx, task, cycle, state)
			return
		}

		retryLoop, terminalFailure, skipNextExecute := h.runCycleLoopVerify(parentCtx, task, cycle, state)
		if retryLoop {
			// ADR-0028: skipNextExecute ⇒ must not call runCycleLoopExecute (no scrub, no runner).
			skipExecute = skipNextExecute
			continue
		}
		if terminalFailure {
			return
		}

		h.runCycleLoopFinalizeSuccess(parentCtx, task, cycle, state)
		return
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) repoRootForGit(_ context.Context) string {
	return strings.TrimSpace(h.opts.WorkingDir)
}
