package model

import (
	"fmt"

	gitmodel "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	settingsmodel "github.com/AlexsanderHamir/Hamix/pkgs/settings/store/model"
	composemodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/store/model"
	"gorm.io/gorm"
)

// AutoMigrateAll runs GORM AutoMigrate for every store model in FK-safe order.
// Callers (postgres.Migrate, test DB setup) must not pass domain types.
//
//funclogmeasure:skip category=hot-path reason="Schema migration helper; postgres.Migrate traces the boot boundary."
func AutoMigrateAll(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&projectmodel.Project{},
		&Task{},
		&TaskDependency{},
		&TaskEvent{},
		&TaskChecklistItem{},
		&TaskChecklistItemCommand{},
		&TaskChecklistCompletion{},
		&composemodel.TaskDraft{},
		&composemodel.TaskTemplate{},
		&TaskCycle{},
		&TaskCyclePhase{},
		&TaskCycleStreamEvent{},
		&TaskCycleCriteriaReport{},
		&TaskCycleVerifyReport{},
		&TaskCycleCommandRun{},
		&TaskCycleCommit{},
		&projectmodel.ProjectContextItem{},
		&projectmodel.ProjectContextEdge{},
		&TaskContextSnapshot{},
		&settingsmodel.AppSettings{},
		&gitmodel.GitRepository{},
		&gitmodel.GitWorktree{},
		&gitmodel.GitBranch{},
	); err != nil {
		return fmt.Errorf("automigrate store models: %w", err)
	}
	return nil
}
