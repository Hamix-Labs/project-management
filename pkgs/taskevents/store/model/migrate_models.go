package model

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrateTaskEvents runs GORM AutoMigrate for task_events.
//
//funclogmeasure:skip category=hot-path reason="Schema migration helper; postgres.Migrate traces the boot boundary."
func AutoMigrateTaskEvents(db *gorm.DB) error {
	if err := db.AutoMigrate(&TaskEvent{}); err != nil {
		return fmt.Errorf("automigrate task event models: %w", err)
	}
	return nil
}
