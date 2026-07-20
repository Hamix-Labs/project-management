package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectsstore "github.com/AlexsanderHamir/Hamix/pkgs/projects/store"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
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

// CreateProjectContext inserts one context item for a project.
func (a *API) CreateProjectContext(ctx context.Context, projectID string, input projectsstore.CreateProjectContextInput) (domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateProjectContext")
	return a.projects.CreateProjectContext(ctx, projectID, input)
}

// ListProjectContext returns context items for a project, pinned items first.
func (a *API) ListProjectContext(ctx context.Context, projectID string, includeUnpinned bool, limit int) ([]domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectContext")
	return a.projects.ListProjectContext(ctx, projectID, includeUnpinned, limit)
}

// ListProjectContextByIDs returns selected context items in caller order.
func (a *API) ListProjectContextByIDs(ctx context.Context, projectID string, ids []string) ([]domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectContextByIDs")
	return a.projects.ListProjectContextByIDs(ctx, projectID, ids)
}

// CreateProjectContextEdge inserts one relationship between project context nodes.
func (a *API) CreateProjectContextEdge(ctx context.Context, projectID string, input projectsstore.CreateProjectContextEdgeInput) (domain.ProjectContextEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateProjectContextEdge")
	return a.projects.CreateProjectContextEdge(ctx, projectID, input)
}

// ListProjectContextEdges returns context edges for one project.
func (a *API) ListProjectContextEdges(ctx context.Context, projectID string, nodeIDs []string) ([]domain.ProjectContextEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectContextEdges")
	return a.projects.ListProjectContextEdges(ctx, projectID, nodeIDs)
}

// UpdateProjectContextEdge applies a partial patch to one project context edge.
func (a *API) UpdateProjectContextEdge(ctx context.Context, projectID, edgeID string, input projectsstore.UpdateProjectContextEdgeInput) (domain.ProjectContextEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateProjectContextEdge")
	return a.projects.UpdateProjectContextEdge(ctx, projectID, edgeID, input)
}

// DeleteProjectContextEdge removes one relationship between project context nodes.
func (a *API) DeleteProjectContextEdge(ctx context.Context, projectID, edgeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteProjectContextEdge")
	return a.projects.DeleteProjectContextEdge(ctx, projectID, edgeID)
}

// UpdateProjectContext applies a partial patch to one project context item.
func (a *API) UpdateProjectContext(ctx context.Context, projectID, itemID string, input projectsstore.UpdateProjectContextInput) (domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateProjectContext")
	return a.projects.UpdateProjectContext(ctx, projectID, itemID, input)
}

// DeleteProjectContext removes one project context item.
func (a *API) DeleteProjectContext(ctx context.Context, projectID, itemID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteProjectContext")
	return a.projects.DeleteProjectContext(ctx, projectID, itemID)
}

// CreateTaskContextSnapshot inserts an immutable project-context snapshot for a cycle.
func (a *API) CreateTaskContextSnapshot(ctx context.Context, input projectsstore.CreateTaskContextSnapshotInput) (taskcoredomain.TaskContextSnapshot, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateTaskContextSnapshot")
	return a.projects.CreateTaskContextSnapshot(ctx, input)
}

// GetTaskContextSnapshotForCycle returns the context snapshot recorded for a cycle.
func (a *API) GetTaskContextSnapshotForCycle(ctx context.Context, cycleID string) (taskcoredomain.TaskContextSnapshot, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetTaskContextSnapshotForCycle")
	return a.projects.GetTaskContextSnapshotForCycle(ctx, cycleID)
}
