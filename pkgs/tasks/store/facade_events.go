package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	taskeventscontract "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
	taskeventsstore "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// TaskEventsPage is one window of audit events (newest first) plus stable paging metadata.
type TaskEventsPage = taskeventscontract.TaskEventsPage

// ThreadEntriesForDisplay returns the conversation for API/list UI. Re-exported from
// pkgs/taskevents/store so the devsim test harness keeps saying
// store.ThreadEntriesForDisplay unchanged.
func ThreadEntriesForDisplay(ev *domain.TaskEvent) []domain.ResponseThreadEntry {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ThreadEntriesForDisplay")
	return taskeventsstore.ThreadEntriesForDisplay(ev)
}

// AppendTaskEvent appends one task_events row if the task exists.
func (s *Store) AppendTaskEvent(ctx context.Context, taskID string, typ domain.EventType, by domain.Actor, data []byte) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AppendTaskEvent")
	return s.events.AppendTaskEvent(ctx, taskID, typ, by, data)
}

// ListTaskEvents returns audit events for a task in ascending sequence order.
func (s *Store) ListTaskEvents(ctx context.Context, taskID string) ([]domain.TaskEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListTaskEvents")
	return s.events.ListTaskEvents(ctx, taskID)
}

// TaskEventCount returns how many audit rows exist for the task.
func (s *Store) TaskEventCount(ctx context.Context, taskID string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.TaskEventCount")
	return s.events.TaskEventCount(ctx, taskID)
}

// LastEventSeq returns the highest seq for the task, or 0 when there are no events.
func (s *Store) LastEventSeq(ctx context.Context, taskID string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.LastEventSeq")
	return s.events.LastEventSeq(ctx, taskID)
}

// GetTaskEvent returns one task_events row by composite key, or ErrNotFound.
func (s *Store) GetTaskEvent(ctx context.Context, taskID string, seq int64) (*domain.TaskEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetTaskEvent")
	return s.events.GetTaskEvent(ctx, taskID, seq)
}

// ListTaskEventsPageCursor returns events in descending seq using keyset paging.
func (s *Store) ListTaskEventsPageCursor(ctx context.Context, taskID string, limit int, beforeSeq, afterSeq *int64) (*TaskEventsPage, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListTaskEventsPageCursor")
	return s.events.ListTaskEventsPageCursor(ctx, taskID, limit, beforeSeq, afterSeq)
}

// ApprovalPending reports whether an approval is outstanding for the task.
func (s *Store) ApprovalPending(ctx context.Context, taskID string) (bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ApprovalPending")
	return s.events.ApprovalPending(ctx, taskID)
}

// AppendTaskEventResponseMessage appends one message to the event thread.
func (s *Store) AppendTaskEventResponseMessage(ctx context.Context, taskID string, seq int64, text string, by domain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AppendTaskEventResponseMessage")
	return s.events.AppendTaskEventResponseMessage(ctx, taskID, seq, text, by)
}
