package contract

import (
	"context"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// TaskEventStore covers task audit/event timeline reads and response append.
type TaskEventStore interface {
	GetTaskEvent(ctx context.Context, taskID string, seq int64) (*taskeventsdomain.TaskEvent, error)
	ListTaskEvents(ctx context.Context, taskID string) ([]taskeventsdomain.TaskEvent, error)
	ListTaskEventsPageCursor(ctx context.Context, taskID string, limit int, beforeSeq, afterSeq *int64) (*TaskEventsPage, error)
	ApprovalPending(ctx context.Context, taskID string) (bool, error)
	AppendTaskEventResponseMessage(ctx context.Context, taskID string, seq int64, text string, by taskeventsdomain.Actor) error
}

// TaskGetter loads a task row for route guards (404 when missing).
type TaskGetter interface {
	Get(ctx context.Context, id string) (*taskcoredomain.Task, error)
}
