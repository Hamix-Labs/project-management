package migrate

import (
	"fmt"

	gitmodel "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	settingsmodel "github.com/AlexsanderHamir/Hamix/pkgs/settings/store/model"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	composemodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/store/model"
	taskcoremodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	cyclesmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/model"
	eventsmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/model"
	"gorm.io/gorm"
)

// autoMigrateStoreModels runs GORM AutoMigrate for every BC model in FK-safe order.
//
//funclogmeasure:skip category=hot-path reason="Schema introspection helper; called at boot in Migrate."
func autoMigrateStoreModels(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&projectmodel.Project{},
		&taskcoremodel.Task{},
		&taskcoremodel.TaskDependency{},
		&eventsmodel.TaskEvent{},
		&checklistmodel.TaskChecklistItem{},
		&checklistmodel.TaskChecklistItemCommand{},
		&checklistmodel.TaskChecklistCompletion{},
		&composemodel.TaskDraft{},
		&composemodel.TaskTemplate{},
		&cyclesmodel.TaskCycle{},
		&cyclesmodel.TaskCyclePhase{},
		&cyclesmodel.TaskCycleStreamEvent{},
		&cyclesmodel.TaskCycleCriteriaReport{},
		&cyclesmodel.TaskCycleVerifyReport{},
		&cyclesmodel.TaskCycleCommandRun{},
		&cyclesmodel.TaskCycleCommit{},
		&projectmodel.ProjectContextItem{},
		&projectmodel.ProjectContextEdge{},
		&taskcoremodel.TaskContextSnapshot{},
		&settingsmodel.AppSettings{},
		&gitmodel.GitRepository{},
		&gitmodel.GitWorktree{},
		&gitmodel.GitBranch{},
	); err != nil {
		return fmt.Errorf("automigrate store models: %w", err)
	}
	return nil
}
