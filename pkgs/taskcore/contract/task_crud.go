package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// TaskCRUDStore covers task CRUD, listing, stats, retry, gate, and dependency edges.
type TaskCRUDStore interface {
	Get(ctx context.Context, id string) (*domain.Task, error)
	Create(ctx context.Context, in CreateTaskInput, by domain.Actor) (*domain.Task, error)
	Update(ctx context.Context, id string, in UpdateTaskInput, by domain.Actor) (*domain.Task, error)
	Delete(ctx context.Context, id string, by domain.Actor) ([]string, error)
	ListFlatPage(ctx context.Context, limit, offset int, filter *ListFilter) ([]domain.Task, bool, error)
	ListFlatAfter(ctx context.Context, limit int, afterID string) ([]domain.Task, bool, error)
	TaskStats(ctx context.Context) (TaskStats, error)
	RequestTaskRetry(ctx context.Context, in RequestRetryInput, by domain.Actor) (*domain.Task, error)
	ApplyTaskGateAction(ctx context.Context, taskID, action string, by domain.Actor) (*domain.Task, error)
	ValidateTaskWorktreeBinding(ctx context.Context, projectID *string, worktreeID string) error
	ListTaskDependencies(ctx context.Context, taskID string) ([]domain.DependencyEdge, error)
	AddTaskDependency(ctx context.Context, taskID, dependsOnTaskID string, satisfies domain.DependencySatisfies) error
	RemoveTaskDependency(ctx context.Context, taskID, dependsOnTaskID string) error
}
