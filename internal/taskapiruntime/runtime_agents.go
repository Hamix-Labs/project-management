package taskapiruntime

import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/agentworker"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/taskapiconfig"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func startReadyTaskAgents(ctx context.Context, taskStore *composition.API, hub *realtime.SSEHub, cmd string) (context.CancelFunc, *agents.MemoryQueue, *agentworker.Supervisor, error) {
	slog.Debug("trace", "cmd", cmd, "operation", "taskapi.startReadyTaskAgents")
	qcap := taskapiconfig.UserTaskAgentQueueCap()
	agentQueue := agents.NewMemoryQueue(qcap)
	taskStore.SetReadyTaskNotifier(agentQueue)
	pickupWake := agents.NewPickupWakeScheduler(taskStore, agentQueue)
	taskStore.SetPickupWake(pickupWake)
	if err := pickupWake.Hydrate(ctx); err != nil {
		return nil, nil, nil, err
	}
	iv := agents.ReconcileTickInterval
	slog.Info("ready task agent queue", "cmd", cmd, "operation", "taskapi.agent_queue", "cap", qcap)
	slog.Info("ready task agent reconcile", "cmd", cmd, "operation", "taskapi.agent_reconcile",
		"tick_interval", iv.String())

	reconcileCtx, reconcileCancel := context.WithCancel(ctx)
	go agents.RunReconcileLoop(reconcileCtx, taskStore, agentQueue, iv, nil)

	provisioner := composition.NewWorktreeProvisioner(taskStore, hub)
	taskStore.SetWorktreeProvisioner(provisioner)
	go provisioner.Run(reconcileCtx)

	sup := agentworker.New(ctx, taskStore, agentQueue, hub)
	if err := sup.Start(ctx); err != nil {
		pickupWake.Stop()
		provisioner.Stop()
		reconcileCancel()
		return nil, nil, nil, err
	}
	taskStore.SetCancelRunForTask(sup.CancelRunForTask)
	taskStore.SetQueueDrop(agentQueue.Drop)
	stopAgents := func() {
		pickupWake.Stop()
		provisioner.Stop()
		reconcileCancel()
	}
	return stopAgents, agentQueue, sup, nil
}
