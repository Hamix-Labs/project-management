package execute

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/orchestration"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// RunPhase executes one execute-phase I/O iteration and returns facts for
// DecideExecutePostRun. The harness root applies effects and anchors state.
func (s *Service) RunPhase(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	ports PhasePorts,
) PhaseResult {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.execute.RunPhase",
		"task_id", task.ID, "cycle_id", cycle.ID)

	execPhase, ok := ports.StartPhase(parentCtx, cycle)
	if !ok {
		return PhaseResult{FatalReason: "execute_phase_start_failed"}
	}

	priorBase, err := s.git.PriorCycleBaseSHA(parentCtx, cycle.ID, execPhase.PhaseSeq)
	if err != nil {
		slog.Warn("agent harness prior cycle base lookup failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.execute.RunPhase.prior_cycle_base",
			"cycle_id", cycle.ID, "err", err)
	}
	snap, err := git.CaptureExecuteGitSnapshot(parentCtx, s.git.Repo(), ports.RepoRoot, ports.WorkingDir, priorBase)
	if err != nil {
		slog.Warn("agent harness git snapshot failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.execute.RunPhase.git_snapshot",
			"cycle_id", cycle.ID, "err", err)
		return PhaseResult{FatalReason: "execute_git_snapshot_failed", ExecPhase: execPhase}
	}
	ports.emitProgress(parentCtx, task.ID, cycle.ID, execPhase,
		runner.SetupProgressEvent(runner.ProgressRunStateSetupGit, "Captured git snapshot…"))

	plan, err := ports.PlanRun(parentCtx, task, cycle)
	if err != nil {
		return PhaseResult{FatalReason: "cursor_resume_plan_failed", ExecPhase: execPhase, Snap: snap}
	}
	ports.emitProgress(parentCtx, task.ID, cycle.ID, execPhase,
		runner.SetupProgressEvent(runner.ProgressRunStateSetupPlan, "Planned Cursor session…"))
	reportDir := ports.ReportDir
	if reportDir == "" {
		reportDir = s.reportDir
	}
	if ports.IsFreshOrFallback != nil && ports.IsFreshOrFallback(plan.Mode) {
		if err := reports.ScrubCycleArtifacts(reportDir, cycle.ID); err != nil {
			slog.Error("agent harness scrub cycle artifacts failed", "cmd", calltrace.LogCmd,
				"operation", "agent.harness.execute.RunPhase.scrub_err",
				"cycle_id", cycle.ID, "report_dir", reportDir, "err", err)
			return PhaseResult{FatalReason: "execute_report_scrub_failed", ExecPhase: execPhase, Snap: snap}
		}
	}
	if err := reports.EnsureReportCycleDir(reportDir, cycle.ID); err != nil {
		slog.Error("agent harness ensure report cycle dir failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.execute.RunPhase.ensure_report_dir_err",
			"cycle_id", cycle.ID, "report_dir", reportDir, "err", err)
		return PhaseResult{FatalReason: "execute_report_dir_ensure_failed", ExecPhase: execPhase, Snap: snap}
	}

	result, runErr := ports.Invoke(parentCtx, task, cycle, execPhase, plan)
	if errors.Is(runErr, runner.ErrResumeSession) {
		ports.emitProgress(parentCtx, task.ID, cycle.ID, execPhase,
			runner.SetupProgressEvent(runner.ProgressRunStateRestartResume, "Restarting agent after failed resume…"))
		fallback := ports.PlanFallback(parentCtx, task, cycle)
		result, runErr = ports.Invoke(parentCtx, task, cycle, execPhase, fallback)
	}
	operatorCancelled := false
	if ports.ConsumeOperatorCancel != nil {
		operatorCancelled = ports.ConsumeOperatorCancel()
	}

	out := PhaseResult{
		ExecPhase:         execPhase,
		Result:            result,
		RunErr:            runErr,
		Snap:              snap,
		OperatorCancelled: operatorCancelled,
		StaleRecovery:     errors.Is(runErr, runner.ErrStale),
	}

	if parentCtx.Err() != nil {
		out.PostRunInput = orchestration.ExecutePostRunInput{ContextCancelled: true}
		return out
	}

	if (runErr == nil || out.StaleRecovery) && !operatorCancelled && !snap.Skipped {
		out.IngestAttempted = true
		ports.emitProgress(parentCtx, task.ID, cycle.ID, execPhase,
			runner.SetupProgressEvent(runner.ProgressRunStateSetupIngest, "Indexing commits…"))
		publish := ports.Publish
		out.IngestOutcome, out.IngestErr = s.git.IngestExecuteCommits(parentCtx, task.ID, cycle, execPhase.PhaseSeq, snap, publish)
		if out.IngestErr != nil {
			slog.Warn("agent harness commit ingest error", "cmd", calltrace.LogCmd,
				"operation", "agent.harness.execute.RunPhase.commit_ingest_err",
				"cycle_id", cycle.ID, "err", out.IngestErr)
		}
	}

	if out.IngestAttempted && out.IngestErr == nil && out.IngestOutcome.FailReason == "" {
		out.CommitCount = out.IngestOutcome.CommitCount
	}

	out.PostRunInput = BuildPostRunInput(parentCtx, runErr, operatorCancelled, snap, out.IngestAttempted, out.IngestOutcome, out.IngestErr)
	return out
}
