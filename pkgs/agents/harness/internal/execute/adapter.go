package execute

import (
	"context"
	"errors"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/orchestration"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
)

// MapRunnerOutcome classifies runner.Run errors at the harness I/O boundary.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func MapRunnerOutcome(err error) orchestration.ExecuteRunnerOutcome {
	if err == nil {
		return orchestration.ExecuteRunnerOutcomeOK
	}
	switch {
	case errors.Is(err, runner.ErrTimeout):
		return orchestration.ExecuteRunnerOutcomeTimeout
	case errors.Is(err, runner.ErrNonZeroExit):
		return orchestration.ExecuteRunnerOutcomeNonZeroExit
	case errors.Is(err, runner.ErrInvalidOutput):
		return orchestration.ExecuteRunnerOutcomeInvalidOutput
	default:
		return orchestration.ExecuteRunnerOutcomeError
	}
}

// BuildPostRunInput maps execute post-run facts to orchestration.ExecutePostRunInput.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func BuildPostRunInput(
	parentCtx context.Context,
	runErr error,
	operatorCancelled bool,
	snap git.PhaseSnapshot,
	ingestAttempted bool,
	ingestOutcome git.ExecuteCommitIngestOutcome,
	ingestErr error,
) orchestration.ExecutePostRunInput {
	in := orchestration.ExecutePostRunInput{
		RunnerOutcome:     MapRunnerOutcome(runErr),
		OperatorCancelled: operatorCancelled,
		ContextCancelled:  parentCtx.Err() != nil,
		CommitIngest: orchestration.ExecuteCommitIngestSummary{
			GitSnapshotSkipped: snap.Skipped,
		},
	}
	shouldIngest := runErr == nil && !operatorCancelled && !snap.Skipped
	if shouldIngest {
		in.CommitIngest.IngestAttempted = ingestAttempted
		if ingestAttempted {
			in.CommitIngest.IngestErr = ingestErr != nil
			in.CommitIngest.FailReason = ingestOutcome.FailReason
		}
	}
	return in
}
