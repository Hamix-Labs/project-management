package migrate

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// migrateTasksStatusClosed (schema rev 19) extends chk_tasks_status to include
// closed (operator exit replacing hard delete).
func migrateTasksStatusClosed(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateTasksStatusClosed")
	if db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}
	if err := db.WithContext(ctx).Exec(`ALTER TABLE tasks DROP CONSTRAINT IF EXISTS chk_tasks_status`).Error; err != nil {
		return fmt.Errorf("drop tasks status constraint: %w", err)
	}
	if err := db.WithContext(ctx).Exec(`ALTER TABLE tasks ADD CONSTRAINT chk_tasks_status CHECK (status IN ('ready','running','blocked','review','done','failed','on_hold','closed'))`).Error; err != nil {
		return fmt.Errorf("add tasks status constraint: %w", err)
	}
	return nil
}
