package migrate

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// legacyDefaultProjectContextTag was the default freeform tag before project
// memory was removed (schema rev 22). Kept as a literal so this historical
// migration does not depend on deleted domain constants.
const legacyDefaultProjectContextTag = "General"

// migrateContextKindToTag replaces project_context_items.kind with tag.
// Idempotent: safe when the table is already gone, when kind is already gone,
// and on fresh installs.
func migrateContextKindToTag(ctx context.Context, db *gorm.DB) error {
	slog.Debug("trace", "operation", "postgres.migrateContextKindToTag")
	if !db.Migrator().HasTable("project_context_items") {
		return nil
	}
	hasKind, err := tableHasColumn(ctx, db, "project_context_items", "kind")
	if err != nil {
		return fmt.Errorf("probe kind column: %w", err)
	}
	if !hasKind {
		return nil
	}
	hasTag, err := tableHasColumn(ctx, db, "project_context_items", "tag")
	if err != nil {
		return fmt.Errorf("probe tag column: %w", err)
	}
	if !hasTag {
		if err := db.WithContext(ctx).Exec(
			`ALTER TABLE project_context_items ADD COLUMN tag TEXT NOT NULL DEFAULT ?`,
			legacyDefaultProjectContextTag,
		).Error; err != nil {
			return fmt.Errorf("add tag column: %w", err)
		}
	}
	// Legacy role enums become General; any other freeform kind value is kept as the tag.
	if err := db.WithContext(ctx).Exec(`
UPDATE project_context_items
SET tag = CASE
  WHEN lower(trim(kind)) IN ('note', 'decision', 'constraint', 'handoff') THEN ?
  WHEN trim(kind) = '' THEN ?
  ELSE trim(kind)
END
`, legacyDefaultProjectContextTag, legacyDefaultProjectContextTag).Error; err != nil {
		return fmt.Errorf("backfill tag from kind: %w", err)
	}
	if db.Dialector != nil && db.Dialector.Name() == "postgres" {
		if err := db.WithContext(ctx).Exec(`ALTER TABLE project_context_items DROP COLUMN IF EXISTS kind`).Error; err != nil {
			return fmt.Errorf("drop kind column: %w", err)
		}
		return nil
	}
	// SQLite: recreate without kind when the column still exists after AutoMigrate
	// kept the legacy column (GORM does not drop columns).
	if err := db.WithContext(ctx).Exec(`ALTER TABLE project_context_items DROP COLUMN kind`).Error; err != nil {
		return fmt.Errorf("drop kind column: %w", err)
	}
	return nil
}
