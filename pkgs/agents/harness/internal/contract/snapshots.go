package contract

import (
	"context"

	projectsstore "github.com/AlexsanderHamir/Hamix/pkgs/projects/store"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// SnapshotStore records immutable project-context snapshots per cycle.
type SnapshotStore interface {
	GetTaskContextSnapshotForCycle(ctx context.Context, cycleID string) (taskcoredomain.TaskContextSnapshot, error)
	CreateTaskContextSnapshot(ctx context.Context, input projectsstore.CreateTaskContextSnapshotInput) (taskcoredomain.TaskContextSnapshot, error)
}
