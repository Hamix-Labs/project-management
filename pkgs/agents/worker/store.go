package worker

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
)

// ReadyTaskQueue is the in-process buffer consumed by Worker and Pool.
// Defined here so pkgs/agents can import worker.Store without an import cycle.
type ReadyTaskQueue interface {
	Receive(ctx context.Context) (taskcoredomain.Task, error)
	AckAfterRecv(id string)
}

// QueueStore is the persistence surface for agent dequeue, reconcile, pickup wake,
// and startup sweep paths beyond harness orchestration.
type QueueStore interface {
	taskcorecontract.AgentQueueStore
	cyclescontract.CycleWorkerStore
}

// Store is the persistence contract for the agent worker, reconcile loop, pickup
// wake scheduler, and harness orchestration. Concrete *composition.API satisfies it
// at the taskapi wiring edge.
type Store interface {
	harness.Store
	QueueStore
}
