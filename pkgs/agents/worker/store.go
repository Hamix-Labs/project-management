package worker

import (
	"context"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

// ReadyTaskQueue is the in-process buffer consumed by Worker and Pool.
// Defined here so pkgs/agents can import worker.Store without an import cycle.
type ReadyTaskQueue interface {
	Receive(ctx context.Context) (domain.Task, error)
	AckAfterRecv(id string)
}

// QueueStore is the persistence surface for agent dequeue, reconcile, pickup wake,
// and startup sweep paths beyond harness orchestration.
type QueueStore interface {
	ListReadyTaskQueueCandidates(ctx context.Context, limit int, cursor *store.ReadyTaskQueueCursor) ([]store.ReadyTaskQueueCandidate, error)
	ListRunningCycles(ctx context.Context) ([]domain.TaskCycle, error)
	ListRunningCyclePhases(ctx context.Context) ([]domain.TaskCyclePhase, error)
	ListDeferredReadyPickupTasks(ctx context.Context, limit int) ([]store.DeferredPickup, error)
	AgentPickup(ctx context.Context, taskID string, by domain.Actor) (*store.AgentPickupResult, error)
	ReadyForAgentPickup(ctx context.Context, t *domain.Task, now time.Time) (bool, store.FailedPredicate, error)
	ResolveTaskGitContext(ctx context.Context, worktreeID string) (store.TaskGitContext, error)
}

// Store is the persistence contract for the agent worker, reconcile loop, pickup
// wake scheduler, and harness orchestration. Concrete *store.Store satisfies it
// at the taskapi wiring edge.
type Store interface {
	harness.Store
	QueueStore
}
