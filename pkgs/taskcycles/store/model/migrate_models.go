package model

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrateCycles runs GORM AutoMigrate for cycle tables in FK-safe order.
// Callers must migrate parent tasks (and checklist items for report FKs) first.
//
//funclogmeasure:skip category=hot-path reason="Schema migration helper; postgres.Migrate traces the boot boundary."
func AutoMigrateCycles(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&TaskCycle{},
		&TaskCyclePhase{},
		&TaskCycleStreamEvent{},
		&TaskCycleCriteriaReport{},
		&TaskCycleVerifyReport{},
		&TaskCycleCommandRun{},
		&TaskCycleCommit{},
	); err != nil {
		return fmt.Errorf("automigrate cycle models: %w", err)
	}
	return nil
}
