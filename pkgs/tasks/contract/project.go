package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// ProjectStore covers projects and the project context graph.
type ProjectStore interface {
	CreateProject(ctx context.Context, input CreateProjectInput) (domain.Project, error)
	ListProjects(ctx context.Context, includeArchived bool, limit int) ([]domain.Project, error)
	GetProject(ctx context.Context, id string) (domain.Project, error)
	UpdateProject(ctx context.Context, id string, input UpdateProjectInput) (domain.Project, error)
	DeleteProject(ctx context.Context, id string) error
	CreateProjectContext(ctx context.Context, projectID string, input CreateProjectContextInput) (domain.ProjectContextItem, error)
	ListProjectContext(ctx context.Context, projectID string, includeUnpinned bool, limit int) ([]domain.ProjectContextItem, error)
	ListProjectContextEdges(ctx context.Context, projectID string, nodeIDs []string) ([]domain.ProjectContextEdge, error)
	CreateProjectContextEdge(ctx context.Context, projectID string, input CreateProjectContextEdgeInput) (domain.ProjectContextEdge, error)
	UpdateProjectContextEdge(ctx context.Context, projectID, edgeID string, input UpdateProjectContextEdgeInput) (domain.ProjectContextEdge, error)
	DeleteProjectContextEdge(ctx context.Context, projectID, edgeID string) error
	UpdateProjectContext(ctx context.Context, projectID, itemID string, input UpdateProjectContextInput) (domain.ProjectContextItem, error)
	DeleteProjectContext(ctx context.Context, projectID, itemID string) error
	ListProjectsByRepository(ctx context.Context, repoID string) ([]domain.Project, error)
}
