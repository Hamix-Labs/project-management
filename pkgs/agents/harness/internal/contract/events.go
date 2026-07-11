package contract

import (
	"context"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// EventStore covers harness task audit append and timeline reads.
type EventStore interface {
	AppendTaskEvent(ctx context.Context, taskID string, typ taskeventsdomain.EventType, by taskcoredomain.Actor, data []byte) error
	ListTaskEvents(ctx context.Context, taskID string) ([]taskeventsdomain.TaskEvent, error)
}
