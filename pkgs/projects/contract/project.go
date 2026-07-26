package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
)

// ProjectStore covers project CRUD and repository-scoped listing.
type ProjectStore interface {
	CreateProject(ctx context.Context, input CreateProjectInput) (domain.Project, error)
	ListProjects(ctx context.Context, includeArchived bool, limit int) ([]domain.Project, error)
	GetProject(ctx context.Context, id string) (domain.Project, error)
	UpdateProject(ctx context.Context, id string, input UpdateProjectInput) (domain.Project, error)
	DeleteProject(ctx context.Context, id string) error
	ListProjectsByRepository(ctx context.Context, repoID string) ([]domain.Project, error)
}
