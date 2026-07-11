package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// TaskEventStore covers task audit/event timeline reads and response append.
type TaskEventStore interface {
	GetTaskEvent(ctx context.Context, taskID string, seq int64) (*domain.TaskEvent, error)
	ListTaskEvents(ctx context.Context, taskID string) ([]domain.TaskEvent, error)
	ListTaskEventsPageCursor(ctx context.Context, taskID string, limit int, beforeSeq, afterSeq *int64) (*TaskEventsPage, error)
	ApprovalPending(ctx context.Context, taskID string) (bool, error)
	AppendTaskEventResponseMessage(ctx context.Context, taskID string, seq int64, text string, by domain.Actor) error
}

// TaskGetter loads a task row for route guards (404 when missing).
type TaskGetter interface {
	Get(ctx context.Context, id string) (*domain.Task, error)
}
