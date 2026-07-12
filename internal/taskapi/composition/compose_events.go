package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventscontract "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	taskeventsstore "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store"
)

// ThreadEntriesForDisplay returns the conversation for API/list UI. Re-exported from
// pkgs/taskevents/store so the devsim test harness keeps saying
// taskeventsstore.ThreadEntriesForDisplay unchanged.
func ThreadEntriesForDisplay(ev *taskeventsdomain.TaskEvent) []taskeventsdomain.ResponseThreadEntry {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.taskeventsstore.ThreadEntriesForDisplay")
	return taskeventsstore.ThreadEntriesForDisplay(ev)
}

// AppendTaskEvent appends one task_events row if the task exists.
func (a *API) AppendTaskEvent(ctx context.Context, taskID string, typ taskeventsdomain.EventType, by taskcoredomain.Actor, data []byte) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AppendTaskEvent")
	return a.events.AppendTaskEvent(ctx, taskID, typ, by, data)
}

// ListTaskEvents returns audit events for a task in ascending sequence order.
func (a *API) ListTaskEvents(ctx context.Context, taskID string) ([]taskeventsdomain.TaskEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListTaskEvents")
	return a.events.ListTaskEvents(ctx, taskID)
}

// TaskEventCount returns how many audit rows exist for the task.
func (a *API) TaskEventCount(ctx context.Context, taskID string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.TaskEventCount")
	return a.events.TaskEventCount(ctx, taskID)
}

// LastEventSeq returns the highest seq for the task, or 0 when there are no events.
func (a *API) LastEventSeq(ctx context.Context, taskID string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.LastEventSeq")
	return a.events.LastEventSeq(ctx, taskID)
}

// GetTaskEvent returns one task_events row by composite key, or ErrNotFound.
func (a *API) GetTaskEvent(ctx context.Context, taskID string, seq int64) (*taskeventsdomain.TaskEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetTaskEvent")
	return a.events.GetTaskEvent(ctx, taskID, seq)
}

// ListTaskEventsPageCursor returns events in descending seq using keyset paging.
func (a *API) ListTaskEventsPageCursor(ctx context.Context, taskID string, limit int, beforeSeq, afterSeq *int64) (*taskeventscontract.TaskEventsPage, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListTaskEventsPageCursor")
	return a.events.ListTaskEventsPageCursor(ctx, taskID, limit, beforeSeq, afterSeq)
}

// ApprovalPending reports whether an approval is outstanding for the task.
func (a *API) ApprovalPending(ctx context.Context, taskID string) (bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ApprovalPending")
	return a.events.ApprovalPending(ctx, taskID)
}

// AppendTaskEventResponseMessage appends one message to the event thread.
func (a *API) AppendTaskEventResponseMessage(ctx context.Context, taskID string, seq int64, text string, by taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AppendTaskEventResponseMessage")
	return a.events.AppendTaskEventResponseMessage(ctx, taskID, seq, text, by)
}
