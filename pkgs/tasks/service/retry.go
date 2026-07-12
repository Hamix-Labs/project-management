package service

import (
	"context"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

// RetryStore persists operator retry intent.
type RetryStore interface {
	RequestTaskRetry(ctx context.Context, in taskcorestore.RequestRetryInput, by taskcoredomain.Actor) (*taskcoredomain.Task, error)
}

// RequestTaskRetry records operator retry intent for a failed task.
//
//funclogmeasure:skip category=delegate-already-logs reason="Thin store wrapper; traces emit at store boundary."
func RequestTaskRetry(
	ctx context.Context,
	st RetryStore,
	in taskcorestore.RequestRetryInput,
	by taskcoredomain.Actor,
) (*taskcoredomain.Task, error) {
	return st.RequestTaskRetry(ctx, in, by)
}
