package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

func (s *Store) ListDeferredReadyPickupTasks(ctx context.Context, limit int) ([]DeferredPickup, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListDeferredReadyPickupTasks")
	return s.taskcore.ListDeferredReadyPickupTasks(ctx, limit)
}

func (s *Store) ListReadyTaskQueueCandidates(ctx context.Context, limit int, cursor *ReadyTaskQueueCursor) ([]ReadyTaskQueueCandidate, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListReadyTaskQueueCandidates")
	return s.taskcore.ListReadyTaskQueueCandidates(ctx, limit, cursor)
}

func (s *Store) ListReadyTasksUserCreated(ctx context.Context, limit int, afterID string) ([]domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListReadyTasksUserCreated")
	return s.taskcore.ListReadyTasksUserCreated(ctx, limit, afterID)
}
