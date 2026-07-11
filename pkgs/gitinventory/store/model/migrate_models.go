package model

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrateGit runs GORM AutoMigrate for git inventory models in FK-safe order.
//
//funclogmeasure:skip category=hot-path reason="Schema migration helper; postgres.Migrate traces the boot boundary."
func AutoMigrateGit(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&GitRepository{},
		&GitWorktree{},
		&GitBranch{},
	); err != nil {
		return fmt.Errorf("automigrate git inventory models: %w", err)
	}
	return nil
}
