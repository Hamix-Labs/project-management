package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// TaskReader is the read/list/stats surface for tasks.
type TaskReader interface {
	TaskGetter
	ListFlatPage(ctx context.Context, limit, offset int, filter *ListFilter) ([]domain.Task, bool, error)
	ListFlatAfter(ctx context.Context, limit int, afterID string) ([]domain.Task, bool, error)
	TaskStats(ctx context.Context) (TaskStats, error)
}

// TaskWriter is the create/update/delete surface for tasks.
type TaskWriter interface {
	Create(ctx context.Context, in CreateTaskInput, by domain.Actor) (*domain.Task, error)
	Update(ctx context.Context, id string, in UpdateTaskInput, by domain.Actor) (*domain.Task, error)
	// Delete removes the task and returns the deleted task id(s) for SSE
	// fan-out. Today the slice contains only the deleted root task id
	// (no cascade of children).
	Delete(ctx context.Context, id string, by domain.Actor) ([]string, error)
}

// TaskDepsStore is the dependency-edge surface for tasks.
type TaskDepsStore interface {
	ListTaskDependencies(ctx context.Context, taskID string) ([]domain.DependencyEdge, error)
	AddTaskDependency(ctx context.Context, taskID, dependsOnTaskID string, satisfies domain.DependencySatisfies) error
	RemoveTaskDependency(ctx context.Context, taskID, dependsOnTaskID string) error
}

// TaskOpsStore is the retry/gate/binding-ops surface for tasks.
type TaskOpsStore interface {
	RequestTaskRetry(ctx context.Context, in RequestRetryInput, by domain.Actor) (*domain.Task, error)
	RequestTaskApprove(ctx context.Context, taskID string, by domain.Actor) (*domain.Task, error)
	RequestTaskPolish(ctx context.Context, in RequestPolishInput, by domain.Actor) (*domain.Task, error)
	RequestTaskOpenPR(ctx context.Context, in RequestOpenPRInput, by domain.Actor) (*domain.Task, error)
	ApplyTaskGateAction(ctx context.Context, taskID string, action GateAction, by domain.Actor) (*domain.Task, error)
	ValidateTaskWorktreeBinding(ctx context.Context, projectID *string, worktreeID string) error
	// Close marks the task closed (idempotent). Composition cancels runs first.
	Close(ctx context.Context, id string, by domain.Actor) (*domain.Task, error)
	// Reopen transitions closed → ready (409 if not closed).
	Reopen(ctx context.Context, id string, by domain.Actor) (*domain.Task, error)
}

// TaskCRUDStore composes focused task seams for wiring-edge consumers.
type TaskCRUDStore interface {
	TaskReader
	TaskWriter
	TaskDepsStore
	TaskOpsStore
}
