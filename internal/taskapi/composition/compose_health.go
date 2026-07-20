package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

const DefaultReadyTimeout = taskcorestore.DefaultReadyTimeout

func (a *API) Ping(ctx context.Context) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Ping")
	return a.taskcore.Ping(ctx)
}

func (a *API) Ready(ctx context.Context) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.Ready")
	return a.taskcore.Ready(ctx)
}
