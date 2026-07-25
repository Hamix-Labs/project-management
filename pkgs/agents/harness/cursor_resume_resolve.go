package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/cursorresume"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func (h *Harness) resolveCursorResume(
	ctx context.Context,
	phase cyclesdomain.Phase,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	opts cycleLoopOpts,
	forceFresh bool,
) (CursorResumeDecision, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.resolveCursorResume",
		"cycle_id", cycle.ID, "phase", string(phase), "force_fresh", forceFresh)

	facts := CursorResumeFacts{
		ForceFresh:         forceFresh,
		RetryMode:          retryModeFromCycleMeta(cycle),
		Phase:              phase,
		ResumeNotice:       opts.resumeNotice,
		ReportTampered:     state.verify.reportTampered,
		GitSkipped:         state.git.gitSnap.Skipped,
		HasPostExecuteHead: state.git.postExecuteHeadSHA != "",
		HeadMatchesAnchor:  true,
		WorkingDir:         h.opts.WorkingDir,
	}
	if !forceFresh {
		settings, err := h.store.GetSettings(ctx)
		if err != nil {
			return CursorResumeDecision{}, err
		}
		facts.SessionResumeEnabled = settings.CursorSessionResumeEnabled

		if !state.git.gitSnap.Skipped && state.git.postExecuteHeadSHA != "" {
			current, ok, herr := h.resolveCurrentHeadSHA(ctx, state.git.gitSnap)
			if herr != nil {
				facts.HeadMatchesAnchor = false
			} else if ok {
				facts.HeadMatchesAnchor = strings.EqualFold(strings.TrimSpace(current), strings.TrimSpace(state.git.postExecuteHeadSHA))
			}
		}
	}

	// Decide without session id first so we skip LastSessionID I/O when an
	// earlier gate already denies (matches prior short-circuit order).
	early := DecideCursorResume(facts)
	if early.DenyReason != "" && early.DenyReason != "no_session_id" && early.DenyReason != "workspace_mismatch" {
		return CursorResumeDecision{Mode: early.Mode, DenyReason: early.DenyReason}, nil
	}

	if !forceFresh {
		lookupCycleID := h.sessionLookupCycleID(ctx, cycle, phase, facts.RetryMode, opts)
		sessionPhase := cursorresume.SessionPhaseForResume(phase)
		sessionID, err := h.store.LastSessionID(ctx, lookupCycleID, sessionPhase)
		if err != nil {
			return CursorResumeDecision{}, err
		}
		facts.SessionID = sessionID
	}

	policy := DecideCursorResume(facts)
	if !policy.AllowResume {
		return CursorResumeDecision{Mode: policy.Mode, DenyReason: policy.DenyReason}, nil
	}

	recoveryCtx := h.buildRecoveryContext(phase, task, cycle, state, opts, facts.RetryMode)
	delta := prompt.ComposeRecoveryDelta(recoveryCtx)
	decision := CursorResumeDecision{
		Mode:            CursorResumeContinue,
		ResumeSessionID: strings.TrimSpace(facts.SessionID),
		Prompt:          delta,
		RecoveryKind:    recoveryCtx.Kind,
	}
	state.resume.lastCursorResumeMode = decision.Mode
	logRecoveryCompose(decision)
	return decision, nil
}

//funclogmeasure:skip category=hot-path reason="Pure cycle id routing; resolveCursorResume logs policy outcome."
func (h *Harness) sessionLookupCycleID(
	ctx context.Context,
	cycle *cyclesdomain.TaskCycle,
	phase cyclesdomain.Phase,
	retryMode taskcoredomain.RetryMode,
	opts cycleLoopOpts,
) string {
	if retryMode == taskcoredomain.RetryResume && cycle.ParentCycleID != nil {
		parentID := strings.TrimSpace(*cycle.ParentCycleID)
		if parentID != "" {
			sessionPhase := cursorresume.SessionPhaseForResume(phase)
			childID, err := h.store.LastSessionID(ctx, cycle.ID, sessionPhase)
			if err == nil && strings.TrimSpace(childID) == "" {
				switch {
				case phase == cyclesdomain.PhaseExecute:
					return parentID
				case phase == cyclesdomain.PhaseVerify && opts.continuation != nil && opts.continuation.Entry == resumeEntryVerifyOnly:
					return parentID
				}
			}
		}
	}
	return cycle.ID
}

//funclogmeasure:skip category=hot-path reason="Pure DTO assembly; ComposeRecoveryDelta logs hint metrics."
func (h *Harness) buildRecoveryContext(
	phase cyclesdomain.Phase,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	opts cycleLoopOpts,
	retryMode taskcoredomain.RetryMode,
) prompt.RecoveryContext {
	reportPath := reports.CriteriaReportPath(h.opts.ReportDir, cycle.ID)
	locked := lockedCriterionIDs(state.verify.previouslyPassed)
	expected := activeCriterionIDs(state)
	runKind := runKindFromCycleMeta(cycle)
	kind := h.selectRecoveryKind(phase, state, opts, retryMode, runKind)
	ctx := prompt.RecoveryContext{
		Kind:                kind,
		Phase:               phase,
		CycleID:             cycle.ID,
		AttemptSeq:          cycle.AttemptSeq,
		VerifyAttempt:       state.verify.verifyAttempt,
		ReportPath:          reportPath,
		FailedCriteria:      failedCriteriaFromVerdicts(state.verify.lastFailedVerdicts),
		LockedCriteria:      locked,
		ReportParseErr:      state.verify.reportParseErr,
		ExpectedIDs:         expected,
		InterruptedPhase:    opts.interruptedPhase,
		PriorVerifyFeedback: state.verify.verifyFeedback,
	}
	if runKind == taskcoredomain.PendingKindPolish {
		ctx.Polish = polishNoticeInputFromCycle(cycle, state, opts.knownCommits)
	}
	if bundle := opts.continuation; bundle != nil && kind == prompt.RecoveryOperatorRetryResume {
		ctx.FailureClass = string(bundle.FailureClass)
		ctx.FailureReason = bundle.FailureReason
		ctx.ScopeFiles = append([]string(nil), bundle.ScopeFiles...)
	}
	if kind == prompt.RecoveryCriteriaReportMissing {
		_ = task
	}
	return ctx
}

//funclogmeasure:skip category=hot-path reason="Pure kind selection from in-memory state."
func (h *Harness) selectRecoveryKind(
	phase cyclesdomain.Phase,
	state *processState,
	opts cycleLoopOpts,
	retryMode taskcoredomain.RetryMode,
	runKind taskcoredomain.PendingRunKind,
) prompt.RecoveryKind {
	return cursorresume.SelectRecoveryKind(cursorresume.RecoveryKindInput{
		Phase:             phase,
		VerifyAttempt:     state.verify.verifyAttempt,
		ReportParseErr:    state.verify.reportParseErr,
		RetryMode:         retryMode,
		RunKind:           runKind,
		HasContinuation:   opts.continuation != nil,
		ResumeNotice:      opts.resumeNotice,
		HasFailedVerdicts: len(state.verify.lastFailedVerdicts) > 0,
	})
}

//funclogmeasure:skip category=hot-path reason="Pure verdict to DTO mapping."
func failedCriteriaFromVerdicts(verdicts []criterionVerdict) []prompt.CriterionFailure {
	in := make([]cursorresume.CriterionFailureInput, 0, len(verdicts))
	for _, v := range verdicts {
		in = append(in, cursorresume.CriterionFailureInput{
			ID:        v.ID,
			Passed:    v.Passed,
			Reasoning: v.Reasoning,
			Verifier:  string(v.Verifier),
		})
	}
	return cursorresume.FailedCriteriaFromInputs(in)
}

//funclogmeasure:skip category=hot-path reason="Pure id extraction from locked verdict map."
func lockedCriterionIDs(locked map[string]criterionVerdict) []string {
	ids := make(map[string]struct{}, len(locked))
	for id := range locked {
		ids[id] = struct{}{}
	}
	return cursorresume.LockedCriterionIDs(ids)
}

//funclogmeasure:skip category=hot-path reason="Pure active checklist id list from state."
func activeCriterionIDs(state *processState) []string {
	all := make([]string, 0, len(state.verify.verifySnap.Criteria))
	for _, it := range state.verify.verifySnap.Criteria {
		all = append(all, it.ID)
	}
	passed := make(map[string]struct{}, len(state.verify.previouslyPassed))
	for id := range state.verify.previouslyPassed {
		passed[id] = struct{}{}
	}
	return cursorresume.ActiveCriterionIDs(all, passed)
}

func logRecoveryCompose(decision CursorResumeDecision) {
	attrs := []any{
		"cmd", calltrace.LogCmd,
		"operation", "agent.harness.Harness.cursorResume",
		"cursor_resume_mode", string(decision.Mode),
		"recovery_hint_bytes", len(decision.Prompt),
	}
	if decision.DenyReason != "" {
		attrs = append(attrs, "deny_reason", decision.DenyReason)
	}
	if decision.RecoveryKind != "" {
		attrs = append(attrs, "recovery_hint_kind", string(decision.RecoveryKind),
			"recovery_failed_criteria_count", 0)
	}
	slog.Debug("trace", attrs...)
}
