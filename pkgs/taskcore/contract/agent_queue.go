package contract

import (
	"context"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// FailedPredicate identifies the first worker readiness check that failed.
type FailedPredicate string

const (
	FailedPredicateNone         FailedPredicate = "none"
	FailedPredicateStatus       FailedPredicate = "status"
	FailedPredicatePickup       FailedPredicate = "pickup"
	FailedPredicateGate         FailedPredicate = "gate"
	FailedPredicateDependencies FailedPredicate = "dependencies"
)

// AgentPickupResult is returned when the worker atomically transitions a task to running.
type AgentPickupResult struct {
	Task          *domain.Task
	ConsumedRetry *domain.PendingRetry
}

// ReadyTaskQueueCursor is a keyset cursor for ListReadyTaskQueueCandidates.
type ReadyTaskQueueCursor struct {
	AfterTaskCreatedAt time.Time
	AfterTaskID        string
	AfterEventRowID    int64
}

// ReadyTaskQueueCandidate is one ready task plus scheduling metadata for the agent queue.
type ReadyTaskQueueCandidate struct {
	Task          domain.Task
	TaskCreatedAt time.Time
	EventRowID    int64
}

// DeferredPickup is a ready task with pickup_not_before still in the future.
type DeferredPickup struct {
	ID              string
	PickupNotBefore time.Time
}

// DeferredPickupCursor is a keyset cursor for ListDeferredReadyPickupTasks.
type DeferredPickupCursor struct {
	NotBefore time.Time
	ID        string
}

// TaskGitContext is the resolved filesystem path and branch name for a task binding.
type TaskGitContext struct {
	WorktreeID   string
	BranchID     string
	WorktreePath string
	BranchName   string
}

// AgentQueueStore covers agent dequeue, pickup wake, and readiness checks.
type AgentQueueStore interface {
	ListReadyTaskQueueCandidates(ctx context.Context, limit int, cursor *ReadyTaskQueueCursor) ([]ReadyTaskQueueCandidate, error)
	ListDeferredReadyPickupTasks(ctx context.Context, limit int, after *DeferredPickupCursor) ([]DeferredPickup, error)
	AgentPickup(ctx context.Context, taskID string, by domain.Actor) (*AgentPickupResult, error)
	ReadyForAgentPickup(ctx context.Context, t *domain.Task, now time.Time) (bool, FailedPredicate, error)
	ResolveTaskGitContext(ctx context.Context, worktreeID string) (TaskGitContext, error)
}
