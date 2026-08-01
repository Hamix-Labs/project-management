package migrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	taskmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"gorm.io/gorm"
)

// migrateGlobalDefaultProject (schema rev 28, ADR-0094) consolidates per-repo
// defaults into one global Default, backfills tasks.repository_id, and replaces
// the per-repo unique index with a global-default unique index.
func migrateGlobalDefaultProject(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateGlobalDefaultProject")
	if err := backfillTaskRepositoryIDs(ctx, db); err != nil {
		return fmt.Errorf("backfill task repository_id: %w", err)
	}
	if err := consolidateToGlobalDefaultProject(ctx, db); err != nil {
		return fmt.Errorf("consolidate global default: %w", err)
	}
	if err := remapComposePayloadsToGlobalDefault(ctx, db); err != nil {
		return fmt.Errorf("remap compose payloads: %w", err)
	}
	return ensureGlobalDefaultUniqueIndex(ctx, db)
}

func backfillTaskRepositoryIDs(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.backfillTaskRepositoryIDs")
	// From project when the project is repo-bound.
	if err := db.WithContext(ctx).Exec(`
UPDATE tasks
SET repository_id = (
  SELECT projects.repository_id FROM projects
  WHERE projects.id = tasks.project_id
    AND projects.repository_id IS NOT NULL
    AND projects.repository_id <> ''
)
WHERE (tasks.repository_id IS NULL OR tasks.repository_id = '')
  AND tasks.project_id IS NOT NULL
  AND tasks.project_id <> ''
`).Error; err != nil {
		return err
	}
	// From worktree when still missing.
	return db.WithContext(ctx).Exec(`
UPDATE tasks
SET repository_id = (
  SELECT git_worktrees.repository_id FROM git_worktrees
  WHERE git_worktrees.id = tasks.worktree_id
)
WHERE (tasks.repository_id IS NULL OR tasks.repository_id = '')
  AND tasks.worktree_id IS NOT NULL
  AND tasks.worktree_id <> ''
`).Error
}

func consolidateToGlobalDefaultProject(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.consolidateToGlobalDefaultProject")
	now := time.Now().UTC()
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		survivor, err := pickOrCreateGlobalDefault(ctx, tx, now)
		if err != nil {
			return err
		}
		var defaults []projectmodel.Project
		if err := tx.Where("is_default = ?", true).Find(&defaults).Error; err != nil {
			return err
		}
		oldIDs := make([]string, 0, len(defaults))
		for _, d := range defaults {
			if d.ID == survivor.ID {
				continue
			}
			oldIDs = append(oldIDs, d.ID)
		}
		if len(oldIDs) > 0 {
			if err := tx.Model(&taskmodel.Task{}).
				Where("project_id IN ?", oldIDs).
				Update("project_id", survivor.ID).Error; err != nil {
				return fmt.Errorf("reassign tasks: %w", err)
			}
			if err := tx.Where("id IN ?", oldIDs).Delete(&projectmodel.Project{}).Error; err != nil {
				return fmt.Errorf("delete per-repo defaults: %w", err)
			}
		}
		updates := map[string]any{
			"name":          projectsdomain.DefaultProjectName,
			"description":   "Built-in project for tasks not assigned to a custom project.",
			"status":        projectsdomain.ProjectStatusActive,
			"is_default":    true,
			"repository_id": nil,
			"updated_at":    now,
		}
		return tx.Model(&projectmodel.Project{}).Where("id = ?", survivor.ID).Updates(updates).Error
	})
}

func pickOrCreateGlobalDefault(ctx context.Context, tx *gorm.DB, now time.Time) (projectmodel.Project, error) {
	slog.Debug("trace", "operation", "postgres.pickOrCreateGlobalDefault")
	var existing projectmodel.Project
	err := tx.WithContext(ctx).
		Where("is_default = ? AND (repository_id IS NULL OR repository_id = '')", true).
		Order("created_at ASC").
		First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return projectmodel.Project{}, err
	}
	err = tx.WithContext(ctx).
		Where("is_default = ?", true).
		Order("created_at ASC").
		First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return projectmodel.Project{}, err
	}
	drow := projectsdomain.GlobalDefaultProject(now)
	row := projectmodel.FromDomainProject(drow)
	if err := tx.WithContext(ctx).Create(&row).Error; err != nil {
		return projectmodel.Project{}, err
	}
	return row, nil
}

func remapComposePayloadsToGlobalDefault(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.remapComposePayloadsToGlobalDefault")
	var survivor projectmodel.Project
	err := db.WithContext(ctx).
		Where("is_default = ? AND (repository_id IS NULL OR repository_id = '')", true).
		First(&survivor).Error
	if err != nil {
		return nil
	}
	for _, table := range []string{"task_templates", "task_drafts"} {
		if !db.Migrator().HasTable(table) || !db.Migrator().HasColumn(table, "payload_json") {
			continue
		}
		type row struct {
			ID          string
			PayloadJSON []byte `gorm:"column:payload_json"`
		}
		var rows []row
		if err := db.WithContext(ctx).Table(table).Select("id, payload_json").Find(&rows).Error; err != nil {
			return err
		}
		for _, r := range rows {
			if len(r.PayloadJSON) == 0 {
				continue
			}
			var payload map[string]any
			if err := json.Unmarshal(r.PayloadJSON, &payload); err != nil {
				continue
			}
			pid, _ := payload["project_id"].(string)
			pid = strings.TrimSpace(pid)
			if pid == "" || pid == survivor.ID {
				continue
			}
			var n int64
			if err := db.WithContext(ctx).Model(&projectmodel.Project{}).
				Where("id = ?", pid).Count(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				continue
			}
			payload["project_id"] = survivor.ID
			raw, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			if err := db.WithContext(ctx).Table(table).Where("id = ?", r.ID).
				Update("payload_json", raw).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureGlobalDefaultUniqueIndex(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.ensureGlobalDefaultUniqueIndex")
	if !isPostgres(db) {
		return nil
	}
	if err := db.WithContext(ctx).Exec(`DROP INDEX IF EXISTS idx_projects_repo_default`).Error; err != nil {
		return err
	}
	return db.WithContext(ctx).Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_global_default
ON projects (is_default)
WHERE is_default = true
`).Error
}
