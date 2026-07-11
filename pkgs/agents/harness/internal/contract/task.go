package contract

import (
	"context"

	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// TaskStore covers harness task create/read/update.
type TaskStore interface {
	Create(ctx context.Context, in taskcorecontract.CreateTaskInput, by taskcoredomain.Actor) (*taskcoredomain.Task, error)
	Get(ctx context.Context, id string) (*taskcoredomain.Task, error)
	Update(ctx context.Context, id string, in taskcorecontract.UpdateTaskInput, by taskcoredomain.Actor) (*taskcoredomain.Task, error)
}
