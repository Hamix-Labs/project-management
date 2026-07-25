package agentworker

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"log/slog"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/agentworker/policy"
	"github.com/AlexsanderHamir/Hamix/internal/taskapiconfig"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
)

type instance struct {
	pool       *worker.Pool
	cancelCtx  context.CancelFunc
	doneCh     chan struct{}
	runTimeout time.Duration
	settings   settingsdomain.AppSettings
	runner     runner.Runner
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func instanceSnapshot(inst *instance, version string) *policy.InstanceSnapshot {
	if inst == nil {
		return nil
	}
	snap := &policy.InstanceSnapshot{Settings: inst.settings}
	if inst.runner != nil {
		if version != "" {
			snap.RunnerVersion = version
		} else {
			snap.RunnerVersion = inst.runner.Version()
		}
	}
	return snap
}

func instanceMatchesSettings(inst *instance, cfg settingsdomain.AppSettings, version string) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.instanceMatchesSettings")
	return policy.InstanceMatchesSettings(instanceSnapshot(inst, version), cfg, version)
}

func stopWorkerInstance(inst *instance, reason string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.stopWorkerInstance",
		"reason", reason)
	if inst == nil || inst.cancelCtx == nil {
		return
	}
	inst.cancelCtx()
	deadline := inst.runTimeout + shutdownGraceAfterRunTimeout
	if inst.runTimeout <= 0 {
		deadline = drainNoLimitTimeout
	}
	select {
	case <-inst.doneCh:
		slog.Info("agent worker instance stopped", "cmd", calltrace.LogCmd,
			"operation", "taskapi.agent_worker.stop", "reason", reason)
	case <-time.After(deadline):
		slog.Warn("agent worker instance drain timeout", "cmd", calltrace.LogCmd,
			"operation", "taskapi.agent_worker.stop_timeout",
			"reason", reason, "deadline", deadline.String())
	}
}

func (s *Supervisor) spawnWorkerInstance(ctx context.Context, cfg settingsdomain.AppSettings, r runner.Runner) *instance {
	runTimeout := time.Duration(cfg.MaxRunDurationSeconds) * time.Second
	notifier := newCycleChangeSSEAdapter(s.publisher, s.notifierMetrics)
	taskUpdatedNotifier := newTaskUpdatedSSEAdapter(s.publisher, s.store, s.notifierMetrics)
	progressNotifier := newRunProgressSSEAdapter(s.publisher, agentRunProgressMinInterval, s.notifierMetrics)
	reportDir := taskapiconfig.WorkerReportDir()
	if err := ensureWorkerReportDirWritable(reportDir); err != nil {
		slog.Warn("agent worker report dir not writable; worker will start but verify will fail",
			"cmd", calltrace.LogCmd, "operation", "taskapi.agent_worker.report_dir_not_writable",
			"path", reportDir, "err", err)
	}
	w := worker.NewPool(s.store, s.queue, r, worker.Options{
		RunTimeout:          runTimeout,
		ReportDir:           reportDir,
		Notifier:            notifier,
		TaskUpdatedNotifier: taskUpdatedNotifier,
		ProgressNotifier:    progressNotifier,
		Metrics:             s.metrics,
	}, taskapiconfig.AgentWorkerConcurrency())

	workerCtx, cancelWorker := context.WithCancel(s.parentCtx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := w.Run(workerCtx); err != nil {
			slog.Error("agent worker exited with error", "cmd", calltrace.LogCmd,
				"operation", "taskapi.agent_worker.exit_err", "err", err)
		}
	}()

	return &instance{
		pool: w, cancelCtx: cancelWorker, doneCh: done,
		runTimeout: runTimeout, settings: cfg, runner: r,
	}
}
