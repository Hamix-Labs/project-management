package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// migrateRepoDefaultProjects (schema rev 6, ADR-0042) backfills per-repo default
// projects, reassigns tasks off the legacy global default, and deletes that row.
func migrateRepoDefaultProjects(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateRepoDefaultProjects")
	if err := ensureDefaultProjectUniqueIndex(ctx, db); err != nil {
		return err
	}
	if err := backfillRepoDefaultProjects(ctx, db); err != nil {
		return err
	}
	if err := reassignTasksToRepoDefaultProjects(ctx, db); err != nil {
		return err
	}
	return deleteLegacyGlobalDefaultProject(ctx, db)
}

func ensureDefaultProjectUniqueIndex(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.ensureDefaultProjectUniqueIndex")
	if !isPostgres(db) {
		return nil
	}
	return db.WithContext(ctx).Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_repo_default
ON projects (repository_id)
WHERE is_default = true AND repository_id IS NOT NULL
`).Error
}

func backfillRepoDefaultProjects(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.backfillRepoDefaultProjects")
	var repos []model.GitRepository
	if err := db.WithContext(ctx).Find(&repos).Error; err != nil {
		return err
	}
	for _, repo := range repos {
		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			_, err := ensureDefaultProjectForRepo(ctx, tx, repo.ID, repo.CreatedAt)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

func ensureDefaultProjectForRepo(ctx context.Context, tx *gorm.DB, repoID string, now time.Time) (model.Project, error) {
	slog.Debug("trace", "operation", "postgres.ensureDefaultProjectForRepo")
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		return model.Project{}, fmt.Errorf("repository_id required")
	}
	var existing model.Project
	err := tx.WithContext(ctx).
		Where("repository_id = ? AND is_default = ?", repoID, true).
		First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Project{}, err
	}
	now = now.UTC()
	row := model.Project{
		ID:             uuid.NewString(),
		Name:           domain.DefaultProjectName,
		Description:    "Built-in project for tasks tied to this repository.",
		Status:         domain.ProjectStatusActive,
		ContextSummary: "Default project for this repository.",
		RepositoryID:   &repoID,
		IsDefault:      true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := tx.WithContext(ctx).Create(&row).Error; err != nil {
		return model.Project{}, err
	}
	return row, nil
}

func reassignTasksToRepoDefaultProjects(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.reassignTasksToRepoDefaultProjects")
	legacyID := domain.LegacyGlobalDefaultProjectID
	type taskRow struct {
		ID         string
		WorktreeID *string
	}
	var tasks []taskRow
	err := db.WithContext(ctx).
		Table("tasks").
		Select("id, worktree_id").
		Where("(project_id IS NULL OR project_id = ?) AND worktree_id IS NOT NULL AND worktree_id <> ''", legacyID).
		Find(&tasks).Error
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.WorktreeID == nil {
			continue
		}
		var wt model.GitWorktree
		if err := db.WithContext(ctx).First(&wt, "id = ?", *task.WorktreeID).Error; err != nil {
			continue
		}
		defaultProj, err := ensureDefaultProjectForRepo(ctx, db, wt.RepositoryID, time.Now().UTC())
		if err != nil {
			continue
		}
		if err := db.WithContext(ctx).
			Model(&model.Task{}).
			Where("id = ?", task.ID).
			Update("project_id", defaultProj.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func deleteLegacyGlobalDefaultProject(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.deleteLegacyGlobalDefaultProject")
	legacyID := domain.LegacyGlobalDefaultProjectID
	if err := db.WithContext(ctx).
		Where("project_id = ?", legacyID).
		Delete(&model.ProjectContextEdge{}).Error; err != nil {
		return err
	}
	if err := db.WithContext(ctx).
		Where("project_id = ?", legacyID).
		Delete(&model.ProjectContextItem{}).Error; err != nil {
		return err
	}
	res := db.WithContext(ctx).Delete(&model.Project{}, "id = ?", legacyID)
	return res.Error
}
