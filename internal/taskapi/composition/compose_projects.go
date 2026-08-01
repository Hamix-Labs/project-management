package composition

import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectsstore "github.com/AlexsanderHamir/Hamix/pkgs/projects/store"
)

// EnsureGlobalDefaultProject delegates to the projects bounded context.
var EnsureGlobalDefaultProject = projectsstore.EnsureGlobalDefaultProject

// ListProjectsByRepository returns user projects for a repository plus the global Default.
func (a *API) ListProjectsByRepository(ctx context.Context, repoID string) ([]domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectsByRepository")
	return a.projects.ListProjectsByRepository(ctx, repoID)
}

// GetGlobalDefaultProject returns the single system Default project.
func (a *API) GetGlobalDefaultProject(ctx context.Context) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGlobalDefaultProject")
	return a.projects.GetGlobalDefaultProject(ctx)
}

// EnsureGlobalDefaultProject inserts the system Default when missing.
func (a *API) EnsureGlobalDefaultProject(ctx context.Context) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.EnsureGlobalDefaultProject")
	return a.projects.EnsureGlobalDefaultProject(ctx)
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
