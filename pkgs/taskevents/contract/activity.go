package contract

import (
	"context"
	"time"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// ActivityEvent is one row in the cross-task activity feed
// (GET /tasks/activity). Types are the fixed set: status_changed,
// phase_failed, approval_granted.
type ActivityEvent struct {
	TaskID    string
	Seq       int64
	At        time.Time
	Type      taskeventsdomain.EventType
	By        taskcoredomain.Actor
	Data      []byte
	TaskTitle *string
}

// ListActivityInput is the paginated query for GET /tasks/activity.
type ListActivityInput struct {
	// Limit is 1–200; defaults to 50.
	Limit int
	// Offset is the number of rows to skip (≥0).
	Offset int
	// Since is an optional RFC3339 lower bound on the event at timestamp.
	// Only events at >= Since are returned.
	Since *time.Time
}

// ListActivityResult is returned by ListTaskActivity.
type ListActivityResult struct {
	Total  int64
	Limit  int
	Offset int
	Events []ActivityEvent
}

// TaskActivityStore lists cross-task activity events.
type TaskActivityStore interface {
	ListTaskActivity(ctx context.Context, in ListActivityInput) (ListActivityResult, error)
}
