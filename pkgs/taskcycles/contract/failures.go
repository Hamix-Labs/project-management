package contract

import (
	"context"
	"time"
)

// CycleFailure is one cycle_failed mirror row projected for operators
// (GET /tasks/cycle-failures and /tasks/stats recent_failures).
type CycleFailure struct {
	TaskID     string
	EventSeq   int64
	At         time.Time
	CycleID    string
	AttemptSeq int64
	Status     string
	Reason     string
}

// ListCycleFailuresInput is the paginated query for cycle failures.
type ListCycleFailuresInput struct {
	Limit  int
	Offset int
	Sort   string
}

// ListCycleFailuresResult is returned by ListCycleFailures.
type ListCycleFailuresResult struct {
	Total               int64
	Failures            []CycleFailure
	ReasonSortTruncated bool
}

// CycleFailuresStore lists paginated cycle failure mirror rows.
type CycleFailuresStore interface {
	ListCycleFailures(ctx context.Context, in ListCycleFailuresInput) (ListCycleFailuresResult, error)
}
