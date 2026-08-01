package resume

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

const (
	cancelledByOperatorReason = "cancelled_by_operator"
	verificationFailedReason  = "verification_failed"
)

// ReconstructCheckpoint rebuilds resume state for an in-flight running cycle.
func (s *Service) ReconstructCheckpoint(ctx context.Context, cycle *cyclesdomain.TaskCycle) (Checkpoint, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.resume.ReconstructCheckpoint",
		"cycle_id", cycle.ID)
	var cp Checkpoint
	cp.LockedPasses = map[string]CriterionVerdict{}
	if cycle == nil {
		return cp, errors.New("resume: nil cycle")
	}
	if cycle.Status != cyclesdomain.CycleStatusRunning {
		return cp, fmt.Errorf("resume: cycle status %q is not running", cycle.Status)
	}

	phases, err := s.store.ListPhasesForCycle(ctx, cycle.ID)
	if err != nil {
		return cp, err
	}
	if len(phases) == 0 {
		return cp, errors.New("resume: cycle has no phases")
	}
	last := phases[len(phases)-1]
	for _, p := range phases {
		if p.Status == cyclesdomain.PhaseStatusRunning {
			return cp, fmt.Errorf("resume: phase_seq=%d still running after finalize", p.PhaseSeq)
		}
	}

	switch {
	case isInterruptPhase(last):
		switch last.Phase {
		case cyclesdomain.PhaseExecute:
			cp.Entry = EntryExecute
		case cyclesdomain.PhaseVerify:
			cp.Entry = EntryVerifyOnly
		default:
			return cp, fmt.Errorf("resume: unexpected interrupted phase %q", last.Phase)
		}
	case last.Phase == cyclesdomain.PhaseExecute && last.Status == cyclesdomain.PhaseStatusSucceeded:
		cp.Entry = EntryAfterExecuteSuccess
	default:
		return cp, fmt.Errorf("resume: cannot continue from phase %q status %q", last.Phase, last.Status)
	}

	lockedPasses, err := s.loadLockedPasses(ctx, cycle.ID)
	if err != nil {
		return cp, err
	}
	cp.LockedPasses = lockedPasses

	commits, err := s.loadKnownCommitsForTask(ctx, cycle.TaskID)
	if err != nil {
		return cp, err
	}
	cp.KnownCommits = commits

	return cp, nil
}

// LoadCheckpointFromParent builds a checkpoint from a terminal parent cycle continuation bundle.
func (s *Service) LoadCheckpointFromParent(ctx context.Context, parentCycleID string) (Checkpoint, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.resume.LoadCheckpointFromParent",
		"parent_cycle_id", parentCycleID)
	bundle, err := s.LoadContinuationBundle(ctx, parentCycleID)
	if err != nil {
		return Checkpoint{LockedPasses: map[string]CriterionVerdict{}}, err
	}
	if !bundle.Sufficient {
		return Checkpoint{LockedPasses: map[string]CriterionVerdict{}},
			fmt.Errorf("continuation: insufficient data for parent %s", parentCycleID)
	}
	return bundleToCheckpoint(bundle), nil
}

func (s *Service) loadKnownCommitsForTask(ctx context.Context, taskID string) ([]cyclesdomain.TaskCycleCommit, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.resume.loadKnownCommitsForTask",
		"task_id", taskID)
	return s.store.ListCommitsForTask(ctx, taskID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Service) loadLockedPasses(ctx context.Context, cycleID string) (map[string]CriterionVerdict, error) {
	lockedPasses := map[string]CriterionVerdict{}
	verifyRows, err := s.store.ListVerifyReportsForCycle(ctx, cycleID)
	if err != nil {
		return nil, err
	}
	for _, row := range verifyRows {
		if !row.Verified {
			continue
		}
		if _, ok := lockedPasses[row.CriterionID]; ok {
			continue
		}
		lockedPasses[row.CriterionID] = CriterionVerdict{
			ID:        row.CriterionID,
			Passed:    true,
			Evidence:  "",
			Verifier:  row.VerifierKind,
			Reasoning: row.Reasoning,
		}
	}
	return lockedPasses, nil
}

func isInterruptPhase(p cyclesdomain.TaskCyclePhase) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.resume.isInterruptPhase",
		"phase_seq", p.PhaseSeq, "phase", string(p.Phase), "status", string(p.Status))
	if p.Status != cyclesdomain.PhaseStatusFailed {
		return false
	}
	if p.Summary == nil {
		return false
	}
	return *p.Summary == cyclesdomain.PhaseInterruptReason
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func bundleToCheckpoint(bundle ContinuationBundle) Checkpoint {
	return Checkpoint{
		Entry:        bundle.Entry,
		LockedPasses: bundle.LockedPasses,
		KnownCommits: bundle.Commits,
		Continuation: &bundle,
	}
}
