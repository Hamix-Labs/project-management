package verify

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/cursorresume"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// PhaseCallbacks notify harness when verify phase rows open and close.
type PhaseCallbacks struct {
	OnStarted func(phase *cyclesdomain.TaskCyclePhase)
	OnEnded   func()
}

// RunPipeline opens a verify phase, runs checks, closes the phase, and returns verdicts.
func (s *Service) RunPipeline(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	snap Snapshot,
	lockedPasses map[string]Verdict,
	mirrorDegradedIn bool,
	phaseCB PhaseCallbacks,
) ([]Verdict, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.RunPipeline",
		"task_id", task.ID, "cycle_id", cycle.ID, "enabled", snap.Enabled)
	if !snap.Enabled {
		return nil, nil
	}
	if err := reports.EnsureReportCycleDir(s.reportDir, cycle.ID); err != nil {
		slog.Warn("agent harness ensureReportCycleDir failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.verify.RunPipeline.ensure_err",
			"cycle_id", cycle.ID, "report_dir", s.reportDir, "err", err)
	}

	verifyStarted := s.clock()
	defer func() {
		s.observeDuration(s.clock().Sub(verifyStarted))
	}()

	phase, err := s.store.StartPhase(parentCtx, cycle.ID, cyclesdomain.PhaseVerify, taskcoredomain.ActorAgent)
	if err != nil {
		slog.Warn("agent harness StartPhase(verify) failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.verify.RunPipeline.start_err",
			"cycle_id", cycle.ID, "err", err)
		return nil, fmt.Errorf("start verify phase: %w", err)
	}
	if phaseCB.OnStarted != nil {
		phaseCB.OnStarted(phase)
	}
	s.publish(cycle.TaskID, cycle.ID)

	runCorrelationID := cyclesdomain.RunCorrelationIDFromDetailsJSON(phase.DetailsJSON)
	s.emitSetupProgress(parentCtx, cycle.TaskID, cycle.ID, phase.PhaseSeq,
		runner.SetupProgressEvent(runner.ProgressRunStateSetupStarted, "Running verify checks…"))

	pre, preErr := s.captureIntegritySnapshot(parentCtx)
	if preErr != nil {
		slog.Warn("agent harness pre-verify integrity snapshot failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.verify.RunPipeline.pre_snapshot_err",
			"cycle_id", cycle.ID, "err", preErr)
	}

	attemptSeq := int64(1)
	verdicts, mirrorDegraded, usage, usagePresent, verifyErr := s.runVerifyChecks(parentCtx, task, cycle, phase.PhaseSeq, runCorrelationID, attemptSeq, snap, lockedPasses, mirrorDegradedIn)

	tampered, tamperReason := s.checkIntegrity(parentCtx, cycle.ID, pre, preErr)

	detailsOpts := PhaseDetailsOpts{
		MirrorDegraded: mirrorDegraded,
		Usage:          usage,
		UsagePresent:   usagePresent,
	}

	phaseStatus := cyclesdomain.PhaseStatusSucceeded
	summary := FormatPhaseSummary(snap.Criteria, verdicts, true)
	var details []byte
	if tampered {
		phaseStatus = cyclesdomain.PhaseStatusFailed
		summary = tamperReason
		verifyErr = &TamperedError{Reason: tamperReason}
	} else if verifyErr != nil {
		phaseStatus = cyclesdomain.PhaseStatusFailed
		if hf, ok := cursorresume.AsHardFail(verifyErr); ok {
			summary = hf.Explain()
			details = EncodePhaseDetails(attemptSeq, snap.Criteria, verdicts, detailsOpts)
			details = MergeFailureDetailsIntoPhaseJSON(details, hf.Kind, hf.Message)
		} else {
			summary = FormatPhaseSummary(snap.Criteria, verdicts, false)
			details = EncodePhaseDetails(attemptSeq, snap.Criteria, verdicts, detailsOpts)
		}
	} else {
		details = EncodePhaseDetails(attemptSeq, snap.Criteria, verdicts, detailsOpts)
	}
	if _, err := s.store.CompletePhase(parentCtx, cyclescontract.CompletePhaseInput{
		CycleID:  cycle.ID,
		PhaseSeq: phase.PhaseSeq,
		Status:   phaseStatus,
		Summary:  &summary,
		Details:  details,
		By:       taskcoredomain.ActorAgent,
	}); err != nil {
		slog.Warn("agent harness CompletePhase(verify) failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.verify.RunPipeline.complete_err",
			"cycle_id", cycle.ID, "phase_seq", phase.PhaseSeq, "err", err)
	}
	if phaseCB.OnEnded != nil {
		phaseCB.OnEnded()
	}
	s.publish(cycle.TaskID, cycle.ID)
	return verdicts, verifyErr
}
