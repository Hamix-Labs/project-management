package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/store/internal"
	taskdomain "github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"gorm.io/gorm"
)

// Store is the GORM-backed persistence facade for projects and project context.
type Store struct {
	db *gorm.DB
}

// NewStore returns a Store backed by db.
func NewStore(db *gorm.DB) *Store {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "projects.store.NewStore")
	return &Store{db: db}
}

// CreateProjectInput is the store input for creating a project.
type CreateProjectInput = internal.CreateProjectInput

// UpdateProjectInput is a partial patch for project metadata.
type UpdateProjectInput = internal.UpdateProjectInput

// CreateProjectContextInput is the store input for appending a project context item.
type CreateProjectContextInput = internal.CreateContextInput

// UpdateProjectContextInput is a partial patch for a project context item.
type UpdateProjectContextInput = internal.UpdateContextInput

// CreateProjectContextEdgeInput is the store input for connecting context nodes.
type CreateProjectContextEdgeInput = internal.CreateContextEdgeInput

// UpdateProjectContextEdgeInput is a partial patch for a project context edge.
type UpdateProjectContextEdgeInput = internal.UpdateContextEdgeInput

// CreateTaskContextSnapshotInput records the rendered project context passed to a cycle.
type CreateTaskContextSnapshotInput = internal.CreateSnapshotInput

// ListProjectsByRepository returns projects tied to a repository.
func (s *Store) ListProjectsByRepository(ctx context.Context, repoID string) ([]domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectsByRepository")
	return internal.ListProjectsByRepository(ctx, s.db, repoID)
}

// GetDefaultProjectForRepository returns the system default project for a repo.
func (s *Store) GetDefaultProjectForRepository(ctx context.Context, repoID string) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetDefaultProjectForRepository")
	return internal.GetDefaultProjectForRepository(ctx, s.db, repoID)
}

// CreateDefaultProjectForRepo inserts the non-deletable default for a newly registered repo.
func CreateDefaultProjectForRepo(ctx context.Context, tx *gorm.DB, repoID string, now time.Time) (domain.Project, error) {
	return internal.CreateDefaultProjectForRepo(ctx, tx, repoID, now)
}

// CreateProject inserts a new active project.
func (s *Store) CreateProject(ctx context.Context, input CreateProjectInput) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateProject")
	return internal.CreateProject(ctx, s.db, input)
}

// ListProjects returns projects ordered by most recently updated first.
func (s *Store) ListProjects(ctx context.Context, includeArchived bool, limit int) ([]domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjects")
	return internal.ListProjects(ctx, s.db, includeArchived, limit)
}

// GetProject returns one project by id.
func (s *Store) GetProject(ctx context.Context, id string) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetProject")
	return internal.GetProject(ctx, s.db, id)
}

// UpdateProject applies a partial project metadata patch.
func (s *Store) UpdateProject(ctx context.Context, id string, input UpdateProjectInput) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateProject")
	return internal.UpdateProject(ctx, s.db, id, input)
}

// DeleteProject removes a project when no tasks still reference it.
func (s *Store) DeleteProject(ctx context.Context, id string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteProject")
	return internal.DeleteProject(ctx, s.db, id)
}

// CreateProjectContext inserts one context item for a project.
func (s *Store) CreateProjectContext(ctx context.Context, projectID string, input CreateProjectContextInput) (domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateProjectContext")
	return internal.CreateContext(ctx, s.db, projectID, input)
}

// ListProjectContext returns context items for a project, pinned items first.
func (s *Store) ListProjectContext(ctx context.Context, projectID string, includeUnpinned bool, limit int) ([]domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectContext")
	return internal.ListContext(ctx, s.db, projectID, includeUnpinned, limit)
}

// ListProjectContextByIDs returns selected context items in caller order.
func (s *Store) ListProjectContextByIDs(ctx context.Context, projectID string, ids []string) ([]domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectContextByIDs")
	return internal.ListContextByIDs(ctx, s.db, projectID, ids)
}

// CreateProjectContextEdge inserts one relationship between project context nodes.
func (s *Store) CreateProjectContextEdge(ctx context.Context, projectID string, input CreateProjectContextEdgeInput) (domain.ProjectContextEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateProjectContextEdge")
	return internal.CreateContextEdge(ctx, s.db, projectID, input)
}

// ListProjectContextEdges returns context edges for one project.
func (s *Store) ListProjectContextEdges(ctx context.Context, projectID string, nodeIDs []string) ([]domain.ProjectContextEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListProjectContextEdges")
	return internal.ListContextEdges(ctx, s.db, projectID, nodeIDs)
}

// UpdateProjectContextEdge applies a partial patch to one project context edge.
func (s *Store) UpdateProjectContextEdge(ctx context.Context, projectID, edgeID string, input UpdateProjectContextEdgeInput) (domain.ProjectContextEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateProjectContextEdge")
	return internal.UpdateContextEdge(ctx, s.db, projectID, edgeID, input)
}

// DeleteProjectContextEdge removes one relationship between project context nodes.
func (s *Store) DeleteProjectContextEdge(ctx context.Context, projectID, edgeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteProjectContextEdge")
	return internal.DeleteContextEdge(ctx, s.db, projectID, edgeID)
}

// UpdateProjectContext applies a partial patch to one project context item.
func (s *Store) UpdateProjectContext(ctx context.Context, projectID, itemID string, input UpdateProjectContextInput) (domain.ProjectContextItem, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateProjectContext")
	return internal.UpdateContext(ctx, s.db, projectID, itemID, input)
}

// DeleteProjectContext removes one project context item.
func (s *Store) DeleteProjectContext(ctx context.Context, projectID, itemID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteProjectContext")
	return internal.DeleteContext(ctx, s.db, projectID, itemID)
}

// CreateTaskContextSnapshot inserts an immutable project-context snapshot for a cycle.
func (s *Store) CreateTaskContextSnapshot(ctx context.Context, input CreateTaskContextSnapshotInput) (taskdomain.TaskContextSnapshot, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateTaskContextSnapshot")
	return internal.CreateSnapshot(ctx, s.db, input)
}

// GetTaskContextSnapshotForCycle returns the context snapshot recorded for a cycle.
func (s *Store) GetTaskContextSnapshotForCycle(ctx context.Context, cycleID string) (taskdomain.TaskContextSnapshot, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetTaskContextSnapshotForCycle")
	return internal.GetSnapshotForCycle(ctx, s.db, cycleID)
}
