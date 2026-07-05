package service

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

// RetryStore persists operator retry intent.
type RetryStore interface {
	RequestTaskRetry(ctx context.Context, in store.RequestRetryInput, by domain.Actor) (*domain.Task, error)
}

// RequestTaskRetry records operator retry intent for a failed task.
//
//funclogmeasure:skip category=delegate-already-logs reason="Thin store wrapper; traces emit at store boundary."
func RequestTaskRetry(
	ctx context.Context,
	st RetryStore,
	in store.RequestRetryInput,
	by domain.Actor,
) (*domain.Task, error) {
	return st.RequestTaskRetry(ctx, in, by)
}
