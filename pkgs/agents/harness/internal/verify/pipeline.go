package verify

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// PhaseCallbacks notify harness when a verify phase row opens and closes.
// Claim-only acceptance does not open a PhaseVerify row; callbacks are unused.
type PhaseCallbacks struct {
	OnStarted func(phase *cyclesdomain.TaskCyclePhase)
	OnEnded   func()
}

// RunPipeline accepts execute claims for all active criteria (including those
// with verify_commands). No worker shell commands and no PhaseVerify Cursor run.
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
	_ = phaseCB
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

	attemptSeq := int64(1)
	verdicts, _, _, _, verifyErr := s.runVerifyChecks(
		parentCtx, task, cycle, 0, "", attemptSeq, snap, lockedPasses, mirrorDegradedIn,
	)
	return verdicts, verifyErr
}
