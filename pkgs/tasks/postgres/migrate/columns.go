package migrate

import (
	"context"
	"log/slog"

	"gorm.io/gorm"
)

func postgresTableHasColumn(ctx context.Context, db *gorm.DB, table, column string) (bool, error) {
	slog.Debug("trace", "operation", "migrate.postgresTableHasColumn", "table", table, "column", column)
	var n int64
	err := db.WithContext(ctx).Raw(`
SELECT COUNT(*) FROM information_schema.columns
 WHERE table_schema = CURRENT_SCHEMA()
   AND table_name = ?
   AND column_name = ?`, table, column).Scan(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// tableHasColumnPortable checks for a column using GORM's Migrator, works
// with both Postgres and SQLite. Used by backfill migrations that must
// become no-ops after contract columns are dropped.
//
//funclogmeasure:skip category=hot-path reason="Schema introspection helper; called at boot in Migrate."
func tableHasColumnPortable(db *gorm.DB, table, column string) bool {
	return db.Migrator().HasColumn(table, column)
}
