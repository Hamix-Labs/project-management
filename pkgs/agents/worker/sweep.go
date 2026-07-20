package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// InterruptPhaseReason is the phase summary recorded when a running phase
// is closed by startup finalization after process interruption. Pinned so
// audit consumers and resume logic can distinguish restart cleanup from
// in-band failures.
const InterruptPhaseReason = cyclesdomain.PhaseInterruptReason

// SweepReason is kept as an alias for callers and tests that referenced
// the historical orphan-sweep constant.
const SweepReason = InterruptPhaseReason

// FinalizeResult is the structured outcome of one FinalizeInterruptedPhases
// call. The counts are best-effort aggregates over rows actually mutated.
type FinalizeResult struct {
	// PhasesFinalized is the number of task_cycle_phases rows successfully
	// flipped from running to failed with InterruptPhaseReason.
	PhasesFinalized int
}

// SweepResult is retained for compatibility with existing log fields and
// tests during the transition away from fail-all orphan sweep.
type SweepResult struct {
	PhasesFailed  int
	CyclesAborted int
	TasksFailed   int
}

// FinalizeInterruptedPhases closes any phase rows left in status='running'
// by a previous process without aborting cycles or failing tasks. Running
// cycles stay running so Harness.Resume can continue the same attempt.
//
// Idempotent: re-running on a clean DB is a no-op. Non-NotFound CompletePhase
// errors are joined and returned so startup can fail closed (B-36 / F-ERR-9).
func FinalizeInterruptedPhases(ctx context.Context, st Store) (FinalizeResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.worker.FinalizeInterruptedPhases")
	var res FinalizeResult
	if st == nil {
		return res, errors.New("agent worker finalize: nil store")
	}

	phases, err := st.ListRunningCyclePhases(ctx)
	if err != nil {
		return res, err
	}
	var errs []error
	for _, p := range phases {
		if ctx.Err() != nil {
			return res, errors.Join(append(errs, ctx.Err())...)
		}
		summary := InterruptPhaseReason
		if _, err := st.CompletePhase(ctx, cyclescontract.CompletePhaseInput{
			CycleID:  p.CycleID,
			PhaseSeq: p.PhaseSeq,
			Status:   cyclesdomain.PhaseStatusFailed,
			Summary:  &summary,
			By:       taskcoredomain.ActorAgent,
		}); err != nil {
			if errors.Is(err, taskcoredomain.ErrNotFound) {
				slog.Info("agent worker finalize CompletePhase not found",
					"cmd", calltrace.LogCmd, "operation", "agent.worker.FinalizeInterruptedPhases.complete_err",
					"cycle_id", p.CycleID, "phase_seq", p.PhaseSeq, "err", err)
				continue
			}
			slog.Error("agent worker finalize CompletePhase failed",
				"cmd", calltrace.LogCmd, "operation", "agent.worker.FinalizeInterruptedPhases.complete_err",
				"cycle_id", p.CycleID, "phase_seq", p.PhaseSeq, "err", err)
			errs = append(errs, fmt.Errorf("complete phase cycle=%s seq=%d: %w", p.CycleID, p.PhaseSeq, err))
			continue
		}
		res.PhasesFinalized++
	}

	slog.Info("agent worker startup finalize complete", "cmd", calltrace.LogCmd,
		"operation", "agent.worker.FinalizeInterruptedPhases.summary",
		"phases_finalized", res.PhasesFinalized, "errors", len(errs))
	if len(errs) > 0 {
		return res, errors.Join(errs...)
	}
	return res, nil
}
