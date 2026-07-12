package main

import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/taskapiconfig"
)

func maybeRunGitReconcileOnStartup(ctx context.Context, taskStore *composition.API) {
	mode := taskapiconfig.GitReconcileOnStartupMode()
	if mode == "" {
		return
	}
	slog.Info("git startup reconcile enabled", "cmd", cmdName, "operation", "taskapi.git_startup_reconcile", "mode", mode)
	taskStore.ReconcileGitRepositoriesOnStartup(ctx, nil)
}
