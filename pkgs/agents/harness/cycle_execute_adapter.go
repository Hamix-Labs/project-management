package harness

import (
	"context"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/execute"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/orchestration"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func mapRunnerOutcome(err error) orchestration.ExecuteRunnerOutcome {
	return execute.MapRunnerOutcome(err)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func buildExecutePostRunInput(
	parentCtx context.Context,
	runErr error,
	operatorCancelled bool,
	snap git.PhaseSnapshot,
	ingestAttempted bool,
	ingestOutcome executeCommitIngestOutcome,
	ingestErr error,
) orchestration.ExecutePostRunInput {
	return execute.BuildPostRunInput(parentCtx, runErr, operatorCancelled, snap, ingestAttempted, ingestOutcome, ingestErr)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func executePhaseStatusFromEffects(effects orchestration.ExecuteEffects) cyclesdomain.PhaseStatus {
	if effects.ContinueToClaimAcceptance {
		return cyclesdomain.PhaseStatusSucceeded
	}
	return cyclesdomain.PhaseStatusFailed
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func overlayOperatorCancelOnResult(result runner.Result, operatorCancelled bool, effects orchestration.ExecuteEffects) runner.Result {
	if !operatorCancelled {
		return result
	}
	if result.Summary == "" || strings.HasPrefix(result.Summary, "cursor: timeout") {
		summary := effects.ResultSummary
		if summary == "" {
			summary = "cancelled by operator"
		}
		result.Summary = summary
	}
	return result
}
