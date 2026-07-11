package model

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrateChecklist runs GORM AutoMigrate for checklist tables in FK-safe order.
//
//funclogmeasure:skip category=hot-path reason="Schema migration helper; postgres.Migrate traces the boot boundary."
func AutoMigrateChecklist(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&TaskChecklistItem{},
		&TaskChecklistItemCommand{},
		&TaskChecklistCompletion{},
	); err != nil {
		return fmt.Errorf("automigrate checklist models: %w", err)
	}
	return nil
}
