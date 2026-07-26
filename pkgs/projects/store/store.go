package store

import (
	"context"
	"log/slog"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/store/internal"
	"gorm.io/gorm"
)

// Store is the GORM-backed persistence facade for projects.
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
func (s *Store) CreateDefaultProjectForRepo(ctx context.Context, repoID string) (domain.Project, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateDefaultProjectForRepo")
	return internal.CreateDefaultProjectForRepo(ctx, s.db, repoID, time.Now().UTC())
}

// CreateDefaultProjectForRepoTx inserts the default project inside an open transaction.
//
//funclogmeasure:skip category=delegate-already-logs reason="Package-level forwarder; internal.CreateDefaultProjectForRepo emits trace at the store chokepoint."
func CreateDefaultProjectForRepo(ctx context.Context, tx *gorm.DB, repoID string, now time.Time) (domain.Project, error) {
	return internal.CreateDefaultProjectForRepo(ctx, tx, repoID, now)
}

// DeleteProjectsForRepository removes all projects for a repository (including defaults).
func (s *Store) DeleteProjectsForRepository(ctx context.Context, repoID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteProjectsForRepository")
	return internal.DeleteProjectsForRepository(ctx, s.db, repoID)
}

// DeleteProjectsForRepositoryTx removes all projects for a repository inside an open transaction.
//
//funclogmeasure:skip category=delegate-already-logs reason="Package-level forwarder; internal.DeleteProjectsForRepository emits trace at the store chokepoint."
func DeleteProjectsForRepository(ctx context.Context, tx *gorm.DB, repoID string) error {
	return internal.DeleteProjectsForRepository(ctx, tx, repoID)
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
