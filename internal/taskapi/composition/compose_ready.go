package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"log/slog"
)

func (a *API) ListDeferredReadyPickupTasks(ctx context.Context, limit int) ([]taskcorecontract.DeferredPickup, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListDeferredReadyPickupTasks")
	return a.taskcore.ListDeferredReadyPickupTasks(ctx, limit)
}

func (a *API) ListReadyTaskQueueCandidates(ctx context.Context, limit int, cursor *taskcorecontract.ReadyTaskQueueCursor) ([]taskcorecontract.ReadyTaskQueueCandidate, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListReadyTaskQueueCandidates")
	return a.taskcore.ListReadyTaskQueueCandidates(ctx, limit, cursor)
}

func (a *API) ListReadyTasksUserCreated(ctx context.Context, limit int, afterID string) ([]taskcoredomain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListReadyTasksUserCreated")
	return a.taskcore.ListReadyTasksUserCreated(ctx, limit, afterID)
}
