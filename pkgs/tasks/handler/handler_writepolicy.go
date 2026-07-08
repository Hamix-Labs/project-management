package handler

import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

// notifyTaskUpdatedEnriched loads the post-commit task row and publishes an
// enriched task_updated frame. Call only after the store mutation succeeds
// (ADR-0026 invariant S1–S2).
func (h *Handler) notifyTaskUpdatedEnriched(ctx context.Context, taskID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.notifyTaskUpdatedEnriched")
	if h.hub == nil {
		return nil
	}
	return realtime.PublishEnrichedTaskUpdated(ctx, h.hub, func(ctx context.Context, id string) (any, error) {
		return h.store.Get(ctx, id)
	}, taskID)
}
