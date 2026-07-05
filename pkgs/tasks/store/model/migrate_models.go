package model

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrateAll runs GORM AutoMigrate for every store model in FK-safe order.
// Callers (postgres.Migrate, test DB setup) must not pass domain types.
//
//funclogmeasure:skip category=hot-path reason="Schema migration helper; postgres.Migrate traces the boot boundary."
func AutoMigrateAll(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&Project{},
		&Task{},
		&TaskDependency{},
		&TaskEvent{},
		&TaskChecklistItem{},
		&TaskChecklistItemCommand{},
		&TaskChecklistCompletion{},
		&TaskDraft{},
		&TaskTemplate{},
		&TaskCycle{},
		&TaskCyclePhase{},
		&TaskCycleStreamEvent{},
		&TaskCycleCriteriaReport{},
		&TaskCycleVerifyReport{},
		&TaskCycleCommandRun{},
		&TaskCycleCommit{},
		&ProjectContextItem{},
		&ProjectContextEdge{},
		&TaskContextSnapshot{},
		&AppSettings{},
		&GitRepository{},
		&GitWorktree{},
		&GitBranch{},
	); err != nil {
		return fmt.Errorf("automigrate store models: %w", err)
	}
	return nil
}
