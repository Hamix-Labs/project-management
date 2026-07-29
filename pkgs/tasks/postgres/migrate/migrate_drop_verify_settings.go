package migrate

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// migrateDropVerifySettingsColumns removes deprecated verify retry and chat-mode
// columns (ADR-0090 command-only verify: one-shot, no operator knobs).
func migrateDropVerifySettingsColumns(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateDropVerifySettingsColumns")
	if db.Dialector == nil {
		return nil
	}
	switch db.Dialector.Name() {
	case "postgres":
		if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP CONSTRAINT IF EXISTS chk_app_settings_verify_chat_mode`).Error; err != nil {
			return fmt.Errorf("drop chk_app_settings_verify_chat_mode: %w", err)
		}
		if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP CONSTRAINT IF EXISTS chk_app_settings_verify_max_retries`).Error; err != nil {
			return fmt.Errorf("drop chk_app_settings_verify_max_retries: %w", err)
		}
		if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP COLUMN IF EXISTS verify_chat_mode`).Error; err != nil {
			return fmt.Errorf("drop app_settings.verify_chat_mode: %w", err)
		}
		if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP COLUMN IF EXISTS verify_max_retries`).Error; err != nil {
			return fmt.Errorf("drop app_settings.verify_max_retries: %w", err)
		}
		if err := db.WithContext(ctx).Exec(`ALTER TABLE tasks DROP COLUMN IF EXISTS verify_chat_mode`).Error; err != nil {
			return fmt.Errorf("drop tasks.verify_chat_mode: %w", err)
		}
	case "sqlite":
		for _, drop := range []struct {
			table  string
			column string
		}{
			{"app_settings", "verify_chat_mode"},
			{"app_settings", "verify_max_retries"},
			{"tasks", "verify_chat_mode"},
		} {
			ok, err := tableHasColumn(ctx, db, drop.table, drop.column)
			if err != nil {
				return fmt.Errorf("probe %s.%s: %w", drop.table, drop.column, err)
			}
			if ok {
				if err := db.WithContext(ctx).Exec(`ALTER TABLE ` + drop.table + ` DROP COLUMN ` + drop.column).Error; err != nil {
					return fmt.Errorf("drop %s.%s: %w", drop.table, drop.column, err)
				}
			}
		}
	}
	return nil
}
