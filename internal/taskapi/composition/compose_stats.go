package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

func (a *API) TaskStats(ctx context.Context) (taskcorestore.TaskStats, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.TaskStats")
	return a.taskcore.TaskStats(ctx)
}

func (a *API) ListCycleFailures(ctx context.Context, in taskcorestore.ListCycleFailuresInput) (taskcorestore.ListCycleFailuresResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListCycleFailures")
	return a.taskcore.ListCycleFailures(ctx, in)
}

func (a *API) CountPreFeatureCycles(ctx context.Context) (taskcorestore.PreFeatureCycleCounts, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CountPreFeatureCycles")
	return a.taskcore.CountPreFeatureCycles(ctx)
}
