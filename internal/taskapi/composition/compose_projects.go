package composition

import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectsstore "github.com/AlexsanderHamir/Hamix/pkgs/projects/store"
)

// CreateDefaultProjectForRepo delegates to the projects bounded context.
// Preferred: call via (*API).projects through CreateGlobalGitRepository composition.
var CreateDefaultProjectForRepo = projectsstore.CreateDefaultProjectForRepo

// ListProjectsByRepository returns projects tied to a repository.
func (a *API) ListProjectsByRepository(ctx context.Context, repoID string) ([]domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectsByRepository")
	return a.projects.ListProjectsByRepository(ctx, repoID)
}

// GetDefaultProjectForRepository returns the system default project for a repo.
func (a *API) GetDefaultProjectForRepository(ctx context.Context, repoID string) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetDefaultProjectForRepository")
	return a.projects.GetDefaultProjectForRepository(ctx, repoID)
}

// CreateProject inserts a new active project.
func (a *API) CreateProject(ctx context.Context, input projectsstore.CreateProjectInput) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateProject")
	return a.projects.CreateProject(ctx, input)
}

// ListProjects returns projects ordered by most recently updated first.
func (a *API) ListProjects(ctx context.Context, includeArchived bool, limit int) ([]domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjects")
	return a.projects.ListProjects(ctx, includeArchived, limit)
}

// GetProject returns one project by id.
func (a *API) GetProject(ctx context.Context, id string) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetProject")
	return a.projects.GetProject(ctx, id)
}

// UpdateProject applies a partial project metadata patch.
func (a *API) UpdateProject(ctx context.Context, id string, input projectsstore.UpdateProjectInput) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateProject")
	return a.projects.UpdateProject(ctx, id, input)
}

// DeleteProject removes a project when no tasks still reference it.
func (a *API) DeleteProject(ctx context.Context, id string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteProject")
	return a.projects.DeleteProject(ctx, id)
}
