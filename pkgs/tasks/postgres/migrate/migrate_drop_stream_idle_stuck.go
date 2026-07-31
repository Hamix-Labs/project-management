package migrate

import (
	"context"
	"log/slog"

	"gorm.io/gorm"
)

// migrateDropStreamIdleStuckColumn is a historical no-op.
//
// Rev 17 dropped app_settings.stream_idle_stuck_seconds. ADR-0091 reintroduces
// the column via AutoMigrate (default 900). Keeping this step as a no-op
// preserves migrate.Run ordering without undoing the re-add.
func migrateDropStreamIdleStuckColumn(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateDropStreamIdleStuckColumn",
		"note", "noop; column reintroduced by ADR-0091 AutoMigrate")
	_ = ctx
	_ = db
	return nil
}
