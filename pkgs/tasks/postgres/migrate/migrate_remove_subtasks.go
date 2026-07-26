package migrate

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// migrateRemoveSubtasks drops the parent/subtask model from upgraded databases.
// Idempotent — safe on fresh installs (no legacy columns) and legacy rows.
func migrateRemoveSubtasks(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateRemoveSubtasks")
	if err := execIfColumnExists(ctx, db, "tasks", "parent_id",
		`UPDATE tasks SET parent_id = NULL WHERE parent_id IS NOT NULL`); err != nil {
		return fmt.Errorf("clear task parent_id: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`UPDATE tasks SET status = 'ready' WHERE status = 'awaiting_subtasks'`).Error; err != nil {
		return fmt.Errorf("promote awaiting_subtasks to ready: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`UPDATE task_dependencies SET satisfies = 'done' WHERE satisfies = 'criteria_complete'`).Error; err != nil {
		return fmt.Errorf("normalize dependency satisfies: %w", err)
	}
	if db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}
	if err := db.WithContext(ctx).Exec(`ALTER TABLE tasks DROP CONSTRAINT IF EXISTS chk_tasks_status`).Error; err != nil {
		return fmt.Errorf("drop tasks status constraint: %w", err)
	}
	// Include closed so re-runs after migrateTasksStatusClosed do not reject existing rows.
	if err := db.WithContext(ctx).Exec(`ALTER TABLE tasks ADD CONSTRAINT chk_tasks_status CHECK (status IN ('ready','running','blocked','review','done','failed','on_hold','closed'))`).Error; err != nil {
		return fmt.Errorf("add tasks status constraint: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`ALTER TABLE task_dependencies DROP CONSTRAINT IF EXISTS chk_task_dependencies_satisfies`).Error; err != nil {
		return fmt.Errorf("drop task_dependencies satisfies constraint: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`ALTER TABLE task_dependencies ADD CONSTRAINT chk_task_dependencies_satisfies CHECK (satisfies IN ('done'))`).Error; err != nil {
		return fmt.Errorf("add task_dependencies satisfies constraint: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`ALTER TABLE tasks DROP COLUMN IF EXISTS parent_id`).Error; err != nil {
		return fmt.Errorf("drop tasks.parent_id: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`ALTER TABLE tasks DROP COLUMN IF EXISTS checklist_inherit`).Error; err != nil {
		return fmt.Errorf("drop tasks.checklist_inherit: %w", err)
	}
	return nil
}

func execIfColumnExists(ctx context.Context, db *gorm.DB, table, column, sql string) error {
	slog.Debug("trace", "operation", "postgres.execIfColumnExists", "table", table, "column", column)
	ok, err := tableHasColumn(ctx, db, table, column)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return db.WithContext(ctx).Exec(sql).Error
}

func tableHasColumn(ctx context.Context, db *gorm.DB, table, column string) (bool, error) {
	slog.Debug("trace", "operation", "postgres.tableHasColumn", "table", table, "column", column)
	if db.Dialector == nil {
		return false, nil
	}
	switch db.Dialector.Name() {
	case "postgres":
		var n int64
		err := db.WithContext(ctx).Raw(
			`SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ? AND column_name = ?`,
			table, column,
		).Scan(&n).Error
		return n > 0, err
	case "sqlite":
		var n int64
		err := db.WithContext(ctx).Raw(
			`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`,
			table, column,
		).Scan(&n).Error
		return n > 0, err
	default:
		return false, fmt.Errorf("unsupported dialector %q for column probe", db.Dialector.Name())
	}
}
