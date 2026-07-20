package verify

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"fmt"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
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
	verifyAttempt int,
	previouslyPassed map[string]Verdict,
	feedback string,
	mirrorDegradedIn bool,
	phaseCB PhaseCallbacks,
) ([]Verdict, string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.RunPipeline",
		"task_id", task.ID, "cycle_id", cycle.ID, "enabled", snap.Enabled)
	if !snap.Enabled {
		return nil, "", nil
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
		return nil, "", fmt.Errorf("start verify phase: %w", err)
	}
	if phaseCB.OnStarted != nil {
		phaseCB.OnStarted(phase)
	}
	s.publish(cycle.TaskID, cycle.ID)

	runCorrelationID := cyclesdomain.RunCorrelationIDFromDetailsJSON(phase.DetailsJSON)

	pre, preErr := s.captureIntegritySnapshot(parentCtx)
	if preErr != nil {
		slog.Warn("agent harness pre-verify integrity snapshot failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.verify.RunPipeline.pre_snapshot_err",
			"cycle_id", cycle.ID, "err", preErr)
	}

	attemptSeq := int64(verifyAttempt) + 1
	verdicts, feedbackOut, mirrorDegraded, verifyErr := s.runVerifyChecks(parentCtx, task, cycle, phase.PhaseSeq, runCorrelationID, attemptSeq, snap, previouslyPassed, feedback, mirrorDegradedIn)

	tampered, tamperReason := s.checkIntegrity(parentCtx, cycle.ID, pre, preErr)

	detailsOpts := PhaseDetailsOpts{
		MirrorDegraded:   mirrorDegraded,
		VerifyRetryCount: verifyAttempt,
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
		summary = FormatPhaseSummary(snap.Criteria, verdicts, false)
		details = EncodePhaseDetails(attemptSeq, snap.Criteria, verdicts, detailsOpts)
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
	return verdicts, feedbackOut, verifyErr
}
