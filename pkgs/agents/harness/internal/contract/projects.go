package contract

import (
	"context"

	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
)

// ProjectStore is the narrow projects surface harness uses for context rendering.
type ProjectStore interface {
	GetProject(ctx context.Context, id string) (projectsdomain.Project, error)
	ListProjectContextByIDs(ctx context.Context, projectID string, ids []string) ([]projectsdomain.ProjectContextItem, error)
	ListProjectContextEdges(ctx context.Context, projectID string, nodeIDs []string) ([]projectsdomain.ProjectContextEdge, error)
}
