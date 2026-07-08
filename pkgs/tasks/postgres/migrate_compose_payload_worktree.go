package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// migrateComposePayloadWorktree (schema rev 7) backfills worktree_id inside
// task_templates and task_drafts compose payloads created before worktree_id
// was required on save.
func migrateComposePayloadWorktree(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateComposePayloadWorktree")
	if err := backfillNamedPayloadWorktreeID(ctx, db, "task_templates"); err != nil {
		return fmt.Errorf("task_templates: %w", err)
	}
	if err := backfillNamedPayloadWorktreeID(ctx, db, "task_drafts"); err != nil {
		return fmt.Errorf("task_drafts: %w", err)
	}
	return nil
}

type namedPayloadRow struct {
	ID          string
	PayloadJSON []byte `gorm:"column:payload_json"`
}

func backfillNamedPayloadWorktreeID(ctx context.Context, db *gorm.DB, table string) error {
	slog.Debug("trace", "operation", "postgres.backfillNamedPayloadWorktreeID", "table", table)
	if !db.Migrator().HasTable(table) {
		return nil
	}
	var rows []namedPayloadRow
	if err := db.WithContext(ctx).Table(table).Select("id, payload_json").Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		patched, changed, err := patchComposePayloadWorktreeID(ctx, db, row.PayloadJSON)
		if err != nil {
			slog.Warn("compose payload worktree backfill skipped", "table", table, "id", row.ID, "err", err)
			continue
		}
		if !changed {
			continue
		}
		if err := db.WithContext(ctx).Table(table).Where("id = ?", row.ID).Update("payload_json", datatypes.JSON(patched)).Error; err != nil {
			return fmt.Errorf("update %s %s: %w", table, row.ID, err)
		}
	}
	return nil
}

func patchComposePayloadWorktreeID(ctx context.Context, db *gorm.DB, raw []byte) (patched []byte, changed bool, err error) {
	slog.Debug("trace", "operation", "postgres.patchComposePayloadWorktreeID")
	if len(raw) == 0 {
		return raw, false, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, false, fmt.Errorf("decode payload: %w", err)
	}
	if wt := strings.TrimSpace(stringFromAny(obj["worktree_id"])); wt != "" {
		return raw, false, nil
	}
	projectID := strings.TrimSpace(stringFromAny(obj["project_id"]))
	proj, repoID, projectRemapped, err := resolveComposeProjectRepository(ctx, db, projectID)
	if err != nil {
		return nil, false, err
	}
	if repoID == "" {
		return raw, false, nil
	}
	worktreeID, err := pickDefaultWorktreeID(ctx, db, repoID)
	if err != nil {
		return nil, false, err
	}
	if worktreeID == "" {
		return raw, false, nil
	}
	obj["worktree_id"] = worktreeID
	if projectRemapped {
		obj["project_id"] = proj.ID
	}
	out, err := json.Marshal(obj)
	if err != nil {
		return nil, false, fmt.Errorf("encode payload: %w", err)
	}
	return out, true, nil
}

func resolveComposeProjectRepository(
	ctx context.Context,
	db *gorm.DB,
	projectID string,
) (proj model.Project, repoID string, remapped bool, err error) {
	slog.Debug("trace", "operation", "postgres.resolveComposeProjectRepository")
	projectID = strings.TrimSpace(projectID)
	if projectID != "" && projectID != domain.LegacyGlobalDefaultProjectID {
		var row model.Project
		loadErr := db.WithContext(ctx).First(&row, "id = ?", projectID).Error
		if loadErr == nil {
			if row.RepositoryID != nil {
				repoID = strings.TrimSpace(*row.RepositoryID)
			}
			return row, repoID, false, nil
		}
		if !errors.Is(loadErr, gorm.ErrRecordNotFound) {
			return model.Project{}, "", false, fmt.Errorf("load project %s: %w", projectID, loadErr)
		}
	}
	// Pre-ADR-0042 templates/drafts may still reference the deleted global default
	// project, or a project row that was removed during repo migration.
	var repos []model.GitRepository
	if err := db.WithContext(ctx).Find(&repos).Error; err != nil {
		return model.Project{}, "", false, err
	}
	if len(repos) != 1 {
		return model.Project{}, "", false, nil
	}
	repoID = strings.TrimSpace(repos[0].ID)
	var defaultProj model.Project
	err = db.WithContext(ctx).
		Where("repository_id = ? AND is_default = ?", repoID, true).
		First(&defaultProj).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Project{}, "", false, nil
		}
		return model.Project{}, "", false, err
	}
	return defaultProj, repoID, true, nil
}

func pickDefaultWorktreeID(ctx context.Context, db *gorm.DB, repositoryID string) (string, error) {
	slog.Debug("trace", "operation", "postgres.pickDefaultWorktreeID")
	repositoryID = strings.TrimSpace(repositoryID)
	if repositoryID == "" {
		return "", nil
	}
	var wt model.GitWorktree
	err := db.WithContext(ctx).
		Where("repository_id = ? AND branch_id IS NOT NULL AND branch_id <> ''", repositoryID).
		Order("is_main DESC, created_at ASC").
		First(&wt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(wt.ID), nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func stringFromAny(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	default:
		return fmt.Sprint(t)
	}
}
