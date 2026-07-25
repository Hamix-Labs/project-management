package agentworker

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"log/slog"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/agentworker/policy"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
)

func (s *Supervisor) probeExecuteRunner(ctx context.Context, cfg settingsdomain.AppSettings) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.agentWorkerSupervisor.probeExecuteRunner",
		"runner", cfg.Runner)
	probeCtx, cancel := context.WithTimeout(ctx, s.probeBudge)
	defer cancel()
	version, _, probeErr := s.probe(probeCtx, cfg.Runner, cfg.CursorBin, s.probeBudge)
	return version, probeErr
}

func (s *Supervisor) runStartupSweep(ctx context.Context) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.agentWorkerSupervisor.runStartupSweep")
	sweepCtx, cancel := context.WithTimeout(ctx, agentWorkerStartupSweepTimeout)
	defer cancel()
	fr, err := worker.FinalizeInterruptedPhases(sweepCtx, s.store)
	if err != nil {
		return err
	}
	slog.Info("agent worker startup finalize ok", "cmd", calltrace.LogCmd,
		"operation", "taskapi.agent_worker.finalize_ok",
		"phases_finalized", fr.PhasesFinalized)
	return nil
}

func (s *Supervisor) probeSchedulingHint(ctx context.Context) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.agentWorkerSupervisor.probeSchedulingHint")
	if s.store == nil {
		return ""
	}
	probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	candidates, err := s.store.ListReadyTaskQueueCandidates(probeCtx, 1, nil)
	if err != nil {
		slog.Debug("scheduling hint: queue probe failed",
			"cmd", calltrace.LogCmd, "operation", "taskapi.agent_worker.scheduling_hint_queue_err",
			"err", err)
		return ""
	}
	queueEmpty := len(candidates) == 0
	if !queueEmpty {
		return ""
	}
	stats, err := s.store.TaskStats(probeCtx)
	if err != nil {
		slog.Debug("scheduling hint: stats probe failed",
			"cmd", calltrace.LogCmd, "operation", "taskapi.agent_worker.scheduling_hint_stats_err",
			"err", err)
		return ""
	}
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.decideSchedulingIdleHint",
		"queue_empty", queueEmpty, "scheduled_count", stats.Scheduled)
	return policy.DecideSchedulingIdleHint(queueEmpty, stats.Scheduled)
}
