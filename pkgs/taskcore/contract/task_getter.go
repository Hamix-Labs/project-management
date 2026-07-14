package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// TaskGetter is the narrow task lookup surface shared by BC handlers and
// agentworker SSE adapters that only need Get for existence checks or enrichment.
type TaskGetter interface {
	Get(ctx context.Context, id string) (*domain.Task, error)
}
