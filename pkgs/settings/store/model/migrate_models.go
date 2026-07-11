package model

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrateSettings runs GORM AutoMigrate for app_settings.
//
//funclogmeasure:skip category=hot-path reason="Schema migration helper; postgres.Migrate traces the boot boundary."
func AutoMigrateSettings(db *gorm.DB) error {
	if err := db.AutoMigrate(&AppSettings{}); err != nil {
		return fmt.Errorf("automigrate app_settings: %w", err)
	}
	return nil
}
