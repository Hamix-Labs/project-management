// package internal owns persistence for first-class projects, curated project
// CRUD.
package internal

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/projects/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateProjectInput is the store input for creating a project.
type CreateProjectInput = contract.CreateProjectInput

// UpdateProjectInput is a partial patch for project metadata.
type UpdateProjectInput = contract.UpdateProjectInput

// CreateProject inserts a new active project.
func CreateProject(ctx context.Context, db *gorm.DB, input CreateProjectInput) (domain.Project, error) {
	defer storekernel.DeferLatency(storekernel.OpCreateProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.CreateProject")
	id := storekernel.ResolveID(input.ID)
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.Project{}, fmt.Errorf("%w: project name required", domain.ErrInvalidInput)
	}
	repoID := trimOptional(input.RepositoryID)
	if repoID == nil {
		return domain.Project{}, fmt.Errorf("%w: repository_id required", domain.ErrInvalidInput)
	}
	if err := ensureRepositoryExists(ctx, db, *repoID); err != nil {
		return domain.Project{}, err
	}
	now := time.Now().UTC()
	drow := domain.Project{
		ID:             id,
		Name:           name,
		Description:    strings.TrimSpace(input.Description),
		Status:         domain.ProjectStatusActive,
		RepositoryID:   repoID,
		NextTaskNumber: 1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	row := projectmodel.FromDomainProject(drow)
	if err := db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Project{}, storekernel.MapWriteError(err, "duplicate project row", domain.ErrConflict, domain.ErrInvalidInput)
	}
	return drow, nil
}

// ListProjects returns projects ordered by most recently updated first.
func ListProjects(ctx context.Context, db *gorm.DB, includeArchived bool, limit int) ([]domain.Project, error) {
	defer storekernel.DeferLatency(storekernel.OpListProjects)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.ListProjects")
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	q := db.WithContext(ctx).Order("updated_at DESC").Limit(limit)
	if !includeArchived {
		q = q.Where("status = ?", domain.ProjectStatusActive)
	}
	var rows []projectmodel.Project
	if err := q.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return projectmodel.ToDomainProjects(rows), nil
}

// GetProject returns one project by id.
func GetProject(ctx context.Context, db *gorm.DB, id string) (domain.Project, error) {
	defer storekernel.DeferLatency(storekernel.OpGetProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.GetProject")
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.Project{}, fmt.Errorf("%w: project id required", domain.ErrInvalidInput)
	}
	var row projectmodel.Project
	if err := db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		return domain.Project{}, storekernel.MapNotFound(err, domain.ErrNotFound)
	}
	return projectmodel.ToDomainProject(row), nil
}

// UpdateProject applies a partial metadata patch and returns the updated row.
func UpdateProject(ctx context.Context, db *gorm.DB, id string, input UpdateProjectInput) (domain.Project, error) {
	defer storekernel.DeferLatency(storekernel.OpUpdateProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.UpdateProject")
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.Project{}, fmt.Errorf("%w: project id required", domain.ErrInvalidInput)
	}
	if err := validateProjectPatch(input); err != nil {
		return domain.Project{}, err
	}
	var out domain.Project
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row projectmodel.Project
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", id).Error; err != nil {
			return storekernel.MapNotFound(err, domain.ErrNotFound)
		}
		drow := projectmodel.ToDomainProject(row)
		if err := validateDefaultProjectPatch(drow, input); err != nil {
			return err
		}
		applyProjectPatch(&drow, input)
		drow.UpdatedAt = time.Now().UTC()
		row = projectmodel.FromDomainProject(drow)
		if err := tx.Save(&row).Error; err != nil {
			return storekernel.MapWriteError(err, "duplicate project row", domain.ErrConflict, domain.ErrInvalidInput)
		}
		out = drow
		return nil
	})
	if err != nil {
		return domain.Project{}, err
	}
	return out, nil
}

// DeleteProject removes a project. Tasks referencing it keep running; project_id
// falls back to NULL via ON DELETE SET NULL on tasks.project_id.
func DeleteProject(ctx context.Context, db *gorm.DB, id string) error {
	defer storekernel.DeferLatency(storekernel.OpDeleteProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.DeleteProject")
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%w: project id required", domain.ErrInvalidInput)
	}
	var row projectmodel.Project
	if err := db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		return storekernel.MapNotFound(err, domain.ErrNotFound)
	}
	if row.IsDefault {
		return fmt.Errorf("%w: default project cannot be deleted", domain.ErrConflict)
	}
	res := db.WithContext(ctx).Delete(&projectmodel.Project{}, "id = ?", id)
	if res.Error != nil {
		return storekernel.MapWriteError(res.Error, "duplicate project row", domain.ErrConflict, domain.ErrInvalidInput)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func validateProjectPatch(input UpdateProjectInput) error {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return fmt.Errorf("%w: project name required", domain.ErrInvalidInput)
	}
	if input.Status != nil {
		switch *input.Status {
		case domain.ProjectStatusActive, domain.ProjectStatusArchived:
		default:
			return fmt.Errorf("%w: invalid project status %q", domain.ErrInvalidInput, *input.Status)
		}
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func validateDefaultProjectPatch(row domain.Project, input UpdateProjectInput) error {
	if !row.IsDefault {
		return nil
	}
	if input.Name != nil && strings.TrimSpace(*input.Name) != domain.DefaultProjectName {
		return fmt.Errorf("%w: default project name cannot be changed", domain.ErrConflict)
	}
	if input.Status != nil && *input.Status != domain.ProjectStatusActive {
		return fmt.Errorf("%w: default project cannot be archived", domain.ErrConflict)
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func applyProjectPatch(row *domain.Project, input UpdateProjectInput) {
	if input.Name != nil {
		row.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		row.Description = strings.TrimSpace(*input.Description)
	}
	if input.Status != nil {
		row.Status = *input.Status
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// CreateDefaultProjectForRepo is removed — use EnsureGlobalDefaultProject (ADR-0094).

// EnsureGlobalDefaultProject inserts the single non-deletable system Default
// when missing. Idempotent: returns the existing global Default when present.
func EnsureGlobalDefaultProject(ctx context.Context, tx *gorm.DB, now time.Time) (domain.Project, error) {
	defer storekernel.DeferLatency(storekernel.OpCreateProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.EnsureGlobalDefaultProject")
	var existing projectmodel.Project
	err := tx.WithContext(ctx).
		Where("is_default = ? AND (repository_id IS NULL OR repository_id = '')", true).
		First(&existing).Error
	if err == nil {
		return projectmodel.ToDomainProject(existing), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Project{}, fmt.Errorf("lookup global default project: %w", err)
	}
	now = now.UTC()
	drow := domain.GlobalDefaultProject(now)
	row := projectmodel.FromDomainProject(drow)
	if err := tx.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Project{}, storekernel.MapWriteError(err, "duplicate default project", domain.ErrConflict, domain.ErrInvalidInput)
	}
	return drow, nil
}

// ListProjectsByRepository returns user projects for a repository plus the
// global system Default (ADR-0094).
func ListProjectsByRepository(ctx context.Context, db *gorm.DB, repoID string) ([]domain.Project, error) {
	defer storekernel.DeferLatency(storekernel.OpListProjects)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.ListProjectsByRepository")
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		return nil, fmt.Errorf("%w: repository_id required", domain.ErrInvalidInput)
	}
	if err := ensureRepositoryExists(ctx, db, repoID); err != nil {
		return nil, err
	}
	var rows []projectmodel.Project
	err := db.WithContext(ctx).
		Where("repository_id = ? OR (is_default = ? AND (repository_id IS NULL OR repository_id = ''))", repoID, true).
		Order("is_default DESC, updated_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list projects by repository: %w", err)
	}
	return projectmodel.ToDomainProjects(rows), nil
}

// ensureRepositoryExists checks git_repositories by string id without importing
// gitinventory models (peer BC cycle break; FK is a plain string column).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ensureRepositoryExists(ctx context.Context, db *gorm.DB, repoID string) error {
	var n int64
	err := db.WithContext(ctx).Table("git_repositories").Where("id = ?", repoID).Limit(1).Count(&n).Error
	if err != nil {
		return err
	}
	if n == 0 {
		return storekernel.MapNotFound(gorm.ErrRecordNotFound, domain.ErrNotFound)
	}
	return nil
}

// GetGlobalDefaultProject returns the single system Default project.
func GetGlobalDefaultProject(ctx context.Context, db *gorm.DB) (domain.Project, error) {
	defer storekernel.DeferLatency(storekernel.OpGetProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.GetGlobalDefaultProject")
	var row projectmodel.Project
	err := db.WithContext(ctx).
		Where("is_default = ? AND (repository_id IS NULL OR repository_id = '')", true).
		First(&row).Error
	if err != nil {
		return domain.Project{}, storekernel.MapNotFound(err, domain.ErrNotFound)
	}
	return projectmodel.ToDomainProject(row), nil
}

// DeleteProjectsForRepository removes user projects tied to a repository.
// The global Default (null repository_id) is never deleted.
func DeleteProjectsForRepository(ctx context.Context, db *gorm.DB, repoID string) error {
	defer storekernel.DeferLatency(storekernel.OpDeleteProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.DeleteProjectsForRepository")
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		return fmt.Errorf("%w: repository_id required", domain.ErrInvalidInput)
	}
	var ids []string
	if err := db.WithContext(ctx).
		Model(&projectmodel.Project{}).
		Where("repository_id = ?", repoID).
		Pluck("id", &ids).Error; err != nil {
		return fmt.Errorf("list projects for repository: %w", err)
	}
	if len(ids) == 0 {
		return nil
	}
	if err := db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&projectmodel.Project{}).Error; err != nil {
		return fmt.Errorf("delete projects: %w", err)
	}
	return nil
}
