package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/cursorresume"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/verify"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// CursorResumeDecision is the harness policy output for one runner.Run.
type CursorResumeDecision struct {
	Mode            CursorResumeMode
	ResumeSessionID string
	Prompt          string
	RecoveryKind    prompt.RecoveryKind
	DenyReason      string
}

func (h *Harness) planExecuteRun(
	ctx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	opts cycleLoopOpts,
) (CursorResumeDecision, error) {
	decision, err := h.resolveCursorResume(ctx, cyclesdomain.PhaseExecute, task, cycle, state, opts, false)
	if err != nil {
		slog.Warn("agent harness cursor resume policy failed; using fresh prompt", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.planExecuteRun.fallback",
			"cycle_id", cycle.ID, "err", err)
		return h.freshExecuteDecision(ctx, task, cycle, state, opts, "policy_error"), nil
	}
	if decision.Mode == CursorResumeFresh {
		decision.Prompt = h.composeExecutePrompt(ctx, task, cycle, state, opts)
	}
	return decision, nil
}

//funclogmeasure:skip category=hot-path reason="Delegates to freshExecuteDecision; resume_fallback logged at invoke site."
func (h *Harness) planExecuteResumeFallback(
	ctx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	opts cycleLoopOpts,
) CursorResumeDecision {
	dec := h.freshExecuteDecision(ctx, task, cycle, state, opts, "resume_failed")
	dec.Mode = CursorResumeFallback
	return dec
}

//funclogmeasure:skip category=hot-path reason="Pure decision struct; composeExecutePrompt logs at invoke site."
func (h *Harness) freshExecuteDecision(
	ctx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	opts cycleLoopOpts,
	denyReason string,
) CursorResumeDecision {
	return CursorResumeDecision{
		Mode:       CursorResumeFresh,
		Prompt:     h.composeExecutePrompt(ctx, task, cycle, state, opts),
		DenyReason: denyReason,
	}
}

func (h *Harness) planVerifyRun(
	ctx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
	snap verificationSnapshot,
	cmdEvidence []verify.CommandEvidence,
	selfReport map[string]reports.CriteriaEntry,
) (verify.VerifyRunPlan, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.planVerifyRun",
		"task_id", task.ID, "cycle_id", cycle.ID)
	opts := cycleLoopOpts{
		resumeNotice:     state.resume.resumeNotice,
		interruptedPhase: state.resume.interruptedPhase,
		continuation:     state.resume.continuation,
	}
	decision, err := h.resolveCursorResume(ctx, cyclesdomain.PhaseVerify, task, cycle, state, opts, false)
	if err != nil {
		return verify.VerifyRunPlan{}, err
	}
	sameChat := state.verify.verifySnap.VerifyChatMode != settingsdomain.VerifyChatModeDifferentChat
	if sameChat && decision.DenyReason == "no_session_id" {
		return verify.VerifyRunPlan{}, cursorresume.MissingSessionForVerify()
	}
	if decision.Mode == CursorResumeFresh || decision.Mode == CursorResumeFallback {
		decision.Prompt = h.verifySvc().BuildVerifyPrompt(ctx, task.ID, snap, cycle.ID, state.verify.lockedPasses, selfReport, cmdEvidence)
	} else {
		rc := h.buildRecoveryContext(cyclesdomain.PhaseVerify, task, cycle, state, opts, retryModeFromCycleMeta(cycle))
		rc.CommandEvidenceDelta = commandEvidenceLines(cmdEvidence)
		rc.VerifyContract = h.verifySvc().BuildVerifyReportContract(
			ctx, task.ID, snap, cycle.ID, state.verify.lockedPasses, selfReport, cmdEvidence,
		)
		decision.Prompt = prompt.ComposeRecoveryDelta(rc)
	}
	state.resume.lastCursorResumeMode = decision.Mode
	logRecoveryCompose(decision)
	return verify.VerifyRunPlan{
		Prompt:           decision.Prompt,
		ResumeSessionID:  decision.ResumeSessionID,
		CursorResumeMode: string(decision.Mode),
		RecoveryKind:     string(decision.RecoveryKind),
	}, nil
}

//funclogmeasure:skip category=hot-path reason="Pure mapping from verify evidence to prompt DTO."
func commandEvidenceLines(evidence []verify.CommandEvidence) []prompt.CommandEvidenceLine {
	out := make([]prompt.CommandEvidenceLine, 0, len(evidence))
	for _, ev := range evidence {
		out = append(out, prompt.CommandEvidenceLine{
			CriterionID: ev.CriterionID,
			Command:     ev.Command,
			ExitCode:    ev.ExitCode,
			Preview:     ev.StdoutPreview,
		})
	}
	return out
}
