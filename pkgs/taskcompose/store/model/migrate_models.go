package model

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrateCompose runs GORM AutoMigrate for task_drafts and task_templates.
//
//funclogmeasure:skip category=hot-path reason="Schema migration helper; postgres.Migrate traces the boot boundary."
func AutoMigrateCompose(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&TaskDraft{},
		&TaskTemplate{},
	); err != nil {
		return fmt.Errorf("automigrate compose models: %w", err)
	}
	return nil
}
