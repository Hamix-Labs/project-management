// Package store implements GORM persistence for task audit events.
package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	taskeventscontract "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/internal/events"
	"gorm.io/gorm"
)

// Store is the GORM-backed persistence facade for task_events reads and thread append.
type Store struct {
	db *gorm.DB
}

// NewStore returns a Store backed by db.
func NewStore(db *gorm.DB) *Store {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.NewStore")
	return &Store{db: db}
}

type (
	// TaskEventsPage is one window of audit events plus paging metadata.
	TaskEventsPage = taskeventscontract.TaskEventsPage
)

// ThreadEntriesForDisplay returns the conversation for API/list UI.
//
//funclogmeasure:skip category=delegate-already-logs reason="Package-level forwarder; events.ThreadEntriesForDisplay emits trace at the store chokepoint."
func ThreadEntriesForDisplay(ev *taskeventsdomain.TaskEvent) []taskeventsdomain.ResponseThreadEntry {
	return events.ThreadEntriesForDisplay(ev)
}

// AppendTaskEvent appends one task_events row if the task exists.
func (s *Store) AppendTaskEvent(ctx context.Context, taskID string, typ taskeventsdomain.EventType, by taskeventsdomain.Actor, data []byte) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.AppendTaskEvent")
	return events.Append(ctx, s.db, taskID, typ, by, data)
}

// ListTaskEvents returns audit events for a task in ascending sequence order.
func (s *Store) ListTaskEvents(ctx context.Context, taskID string) ([]taskeventsdomain.TaskEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.ListTaskEvents")
	return events.List(ctx, s.db, taskID)
}

// TaskEventCount returns how many audit rows exist for the task.
func (s *Store) TaskEventCount(ctx context.Context, taskID string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.TaskEventCount")
	return events.Count(ctx, s.db, taskID)
}

// LastEventSeq returns the highest seq for the task, or 0 when there are no events.
func (s *Store) LastEventSeq(ctx context.Context, taskID string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.LastEventSeq")
	return events.LastSeq(ctx, s.db, taskID)
}

// GetTaskEvent returns one task_events row by composite key, or ErrNotFound.
func (s *Store) GetTaskEvent(ctx context.Context, taskID string, seq int64) (*taskeventsdomain.TaskEvent, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.GetTaskEvent")
	return events.Get(ctx, s.db, taskID, seq)
}

// ListTaskEventsPageCursor returns events in descending seq using keyset paging.
func (s *Store) ListTaskEventsPageCursor(ctx context.Context, taskID string, limit int, beforeSeq, afterSeq *int64) (*TaskEventsPage, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.ListTaskEventsPageCursor")
	return events.PageCursor(ctx, s.db, taskID, limit, beforeSeq, afterSeq)
}

// ApprovalPending reports whether an approval is outstanding for the task.
func (s *Store) ApprovalPending(ctx context.Context, taskID string) (bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.ApprovalPending")
	return events.ApprovalPending(ctx, s.db, taskID)
}

// AppendTaskEventResponseMessage appends one message to the event thread.
func (s *Store) AppendTaskEventResponseMessage(ctx context.Context, taskID string, seq int64, text string, by taskeventsdomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.store.AppendTaskEventResponseMessage")
	return events.AppendResponseMessage(ctx, s.db, taskID, seq, text, by)
}
