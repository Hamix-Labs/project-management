package migrate

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// migrateDropVerifyCommandTimeoutColumn removes the deprecated
// app_settings.verify_command_timeout_seconds column (and its postgres check
// constraint if present). Timeout is now per verify command on
// task_checklist_item_commands.timeout_seconds.
func migrateDropVerifyCommandTimeoutColumn(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateDropVerifyCommandTimeoutColumn")
	if db.Dialector == nil {
		return nil
	}
	switch db.Dialector.Name() {
	case "postgres":
		if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP CONSTRAINT IF EXISTS chk_app_settings_verify_command_timeout_seconds`).Error; err != nil {
			return fmt.Errorf("drop chk_app_settings_verify_command_timeout_seconds: %w", err)
		}
		if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP COLUMN IF EXISTS verify_command_timeout_seconds`).Error; err != nil {
			return fmt.Errorf("drop app_settings.verify_command_timeout_seconds: %w", err)
		}
	case "sqlite":
		ok, err := tableHasColumn(ctx, db, "app_settings", "verify_command_timeout_seconds")
		if err != nil {
			return fmt.Errorf("probe app_settings.verify_command_timeout_seconds: %w", err)
		}
		if ok {
			if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP COLUMN verify_command_timeout_seconds`).Error; err != nil {
				return fmt.Errorf("drop app_settings.verify_command_timeout_seconds: %w", err)
			}
		}
	}
	return nil
}
