package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/prompt"
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
