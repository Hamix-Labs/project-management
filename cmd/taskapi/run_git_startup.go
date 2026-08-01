package main

import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/internal/taskapiconfig"
	"github.com/AlexsanderHamir/Hamix/internal/taskapiruntime"
)

func maybeRunGitReconcileOnStartup(ctx context.Context, rt *taskapiruntime.Runtime) {
	mode := taskapiconfig.GitReconcileOnStartupMode()
	if mode == "" {
		return
	}
	slog.Info("git startup reconcile enabled", "cmd", cmdName, "operation", "taskapi.git_startup_reconcile", "mode", mode)
	rt.ReconcileGitOnStartup(ctx)
}
