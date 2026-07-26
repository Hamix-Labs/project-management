package migrate

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// migrateRemoveProjectContext drops project memory tables and leftover columns.
// Idempotent — safe on fresh installs and upgraded databases.
func migrateRemoveProjectContext(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateRemoveProjectContext")
	if err := db.WithContext(ctx).Exec(`DROP TABLE IF EXISTS task_context_snapshots`).Error; err != nil {
		return fmt.Errorf("drop task_context_snapshots: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`DROP TABLE IF EXISTS project_context_edges`).Error; err != nil {
		return fmt.Errorf("drop project_context_edges: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`DROP TABLE IF EXISTS project_context_items`).Error; err != nil {
		return fmt.Errorf("drop project_context_items: %w", err)
	}
	if err := execIfColumnExists(ctx, db, "tasks", "project_context_item_ids",
		`ALTER TABLE tasks DROP COLUMN project_context_item_ids`); err != nil {
		return fmt.Errorf("drop tasks.project_context_item_ids: %w", err)
	}
	if err := execIfColumnExists(ctx, db, "projects", "context_summary",
		`ALTER TABLE projects DROP COLUMN context_summary`); err != nil {
		return fmt.Errorf("drop projects.context_summary: %w", err)
	}
	return nil
}
