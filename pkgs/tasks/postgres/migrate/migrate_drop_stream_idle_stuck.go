package migrate

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// migrateDropStreamIdleStuckColumn removes the deprecated
// app_settings.stream_idle_stuck_seconds column (and its postgres check
// constraint if present). Stream-idle kill was removed end-to-end;
// runners own their own timeouts.
func migrateDropStreamIdleStuckColumn(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateDropStreamIdleStuckColumn")
	if db.Dialector == nil {
		return nil
	}
	switch db.Dialector.Name() {
	case "postgres":
		if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP CONSTRAINT IF EXISTS chk_app_settings_stream_idle_stuck_seconds`).Error; err != nil {
			return fmt.Errorf("drop chk_app_settings_stream_idle_stuck_seconds: %w", err)
		}
		if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP COLUMN IF EXISTS stream_idle_stuck_seconds`).Error; err != nil {
			return fmt.Errorf("drop app_settings.stream_idle_stuck_seconds: %w", err)
		}
	case "sqlite":
		ok, err := tableHasColumn(ctx, db, "app_settings", "stream_idle_stuck_seconds")
		if err != nil {
			return fmt.Errorf("probe app_settings.stream_idle_stuck_seconds: %w", err)
		}
		if ok {
			if err := db.WithContext(ctx).Exec(`ALTER TABLE app_settings DROP COLUMN stream_idle_stuck_seconds`).Error; err != nil {
				return fmt.Errorf("drop app_settings.stream_idle_stuck_seconds: %w", err)
			}
		}
	}
	return nil
}
