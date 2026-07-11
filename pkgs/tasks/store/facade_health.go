package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

const DefaultReadyTimeout = taskcorestore.DefaultReadyTimeout

func (s *Store) Ping(ctx context.Context) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Ping")
	return s.taskcore.Ping(ctx)
}

func (s *Store) Ready(ctx context.Context) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Ready")
	return s.taskcore.Ready(ctx)
}
