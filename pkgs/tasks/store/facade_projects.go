package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectsstore "github.com/AlexsanderHamir/Hamix/pkgs/projects/store"
	taskdomain "github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// Project store input aliases — re-exported for handlers and harness callers.
type (
	CreateProjectInput             = projectsstore.CreateProjectInput
	UpdateProjectInput             = projectsstore.UpdateProjectInput
	CreateProjectContextInput      = projectsstore.CreateProjectContextInput
	UpdateProjectContextInput      = projectsstore.UpdateProjectContextInput
	CreateProjectContextEdgeInput  = projectsstore.CreateProjectContextEdgeInput
	UpdateProjectContextEdgeInput  = projectsstore.UpdateProjectContextEdgeInput
	CreateTaskContextSnapshotInput = projectsstore.CreateTaskContextSnapshotInput
)

// CreateDefaultProjectForRepo delegates to the projects bounded context.
var CreateDefaultProjectForRepo = projectsstore.CreateDefaultProjectForRepo

// GetDefaultProjectForRepository returns the system default project for a repo.
func (s *Store) GetDefaultProjectForRepository(ctx context.Context, repoID string) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetDefaultProjectForRepository")
	return s.projects.GetDefaultProjectForRepository(ctx, repoID)
}

// CreateProject inserts a new active project.
func (s *Store) CreateProject(ctx context.Context, input CreateProjectInput) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateProject")
	return s.projects.CreateProject(ctx, input)
}

// ListProjects returns projects ordered by most recently updated first.
func (s *Store) ListProjects(ctx context.Context, includeArchived bool, limit int) ([]domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjects")
	return s.projects.ListProjects(ctx, includeArchived, limit)
}

// GetProject returns one project by id.
func (s *Store) GetProject(ctx context.Context, id string) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetProject")
	return s.projects.GetProject(ctx, id)
}

// UpdateProject applies a partial project metadata patch.
func (s *Store) UpdateProject(ctx context.Context, id string, input UpdateProjectInput) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateProject")
	return s.projects.UpdateProject(ctx, id, input)
}

// DeleteProject removes a project when no tasks still reference it.
func (s *Store) DeleteProject(ctx context.Context, id string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteProject")
	return s.projects.DeleteProject(ctx, id)
}

// CreateProjectContext inserts one context item for a project.
func (s *Store) CreateProjectContext(ctx context.Context, projectID string, input CreateProjectContextInput) (domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateProjectContext")
	return s.projects.CreateProjectContext(ctx, projectID, input)
}

// ListProjectContext returns context items for a project, pinned items first.
func (s *Store) ListProjectContext(ctx context.Context, projectID string, includeUnpinned bool, limit int) ([]domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectContext")
	return s.projects.ListProjectContext(ctx, projectID, includeUnpinned, limit)
}

// ListProjectContextByIDs returns selected context items in caller order.
func (s *Store) ListProjectContextByIDs(ctx context.Context, projectID string, ids []string) ([]domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectContextByIDs")
	return s.projects.ListProjectContextByIDs(ctx, projectID, ids)
}

// CreateProjectContextEdge inserts one relationship between project context nodes.
func (s *Store) CreateProjectContextEdge(ctx context.Context, projectID string, input CreateProjectContextEdgeInput) (domain.ProjectContextEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateProjectContextEdge")
	return s.projects.CreateProjectContextEdge(ctx, projectID, input)
}

// ListProjectContextEdges returns context edges for one project.
func (s *Store) ListProjectContextEdges(ctx context.Context, projectID string, nodeIDs []string) ([]domain.ProjectContextEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectContextEdges")
	return s.projects.ListProjectContextEdges(ctx, projectID, nodeIDs)
}

// UpdateProjectContextEdge applies a partial patch to one project context edge.
func (s *Store) UpdateProjectContextEdge(ctx context.Context, projectID, edgeID string, input UpdateProjectContextEdgeInput) (domain.ProjectContextEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateProjectContextEdge")
	return s.projects.UpdateProjectContextEdge(ctx, projectID, edgeID, input)
}

// DeleteProjectContextEdge removes one relationship between project context nodes.
func (s *Store) DeleteProjectContextEdge(ctx context.Context, projectID, edgeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteProjectContextEdge")
	return s.projects.DeleteProjectContextEdge(ctx, projectID, edgeID)
}

// UpdateProjectContext applies a partial patch to one project context item.
func (s *Store) UpdateProjectContext(ctx context.Context, projectID, itemID string, input UpdateProjectContextInput) (domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateProjectContext")
	return s.projects.UpdateProjectContext(ctx, projectID, itemID, input)
}

// DeleteProjectContext removes one project context item.
func (s *Store) DeleteProjectContext(ctx context.Context, projectID, itemID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteProjectContext")
	return s.projects.DeleteProjectContext(ctx, projectID, itemID)
}

// CreateTaskContextSnapshot inserts an immutable project-context snapshot for a cycle.
func (s *Store) CreateTaskContextSnapshot(ctx context.Context, input CreateTaskContextSnapshotInput) (taskdomain.TaskContextSnapshot, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateTaskContextSnapshot")
	return s.projects.CreateTaskContextSnapshot(ctx, input)
}

// GetTaskContextSnapshotForCycle returns the context snapshot recorded for a cycle.
func (s *Store) GetTaskContextSnapshotForCycle(ctx context.Context, cycleID string) (taskdomain.TaskContextSnapshot, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetTaskContextSnapshotForCycle")
	return s.projects.GetTaskContextSnapshotForCycle(ctx, cycleID)
}
