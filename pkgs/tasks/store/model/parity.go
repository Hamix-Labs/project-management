package model

import (
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	settingsmodel "github.com/AlexsanderHamir/Hamix/pkgs/settings/store/model"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	composedomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/domain"
	composemodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcoremodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/model"
	eventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/model"
)

// ParityPair binds a domain struct prototype to its model counterpart for
// schema- and field-parity guards.
type ParityPair struct {
	Name   string
	Domain any
	Model  any
	Table  string
	// ModelMigrateExtra lists additional model structs AutoMigrate must run
	// before the primary model type (e.g. parent tables for association FKs).
	ModelMigrateExtra []any
}

// ParityPairs is the single registry both parity tests iterate.
var ParityPairs = []ParityPair{
	{
		Name:   "AppSettings",
		Domain: &settingsdomain.AppSettings{},
		Model:  &settingsmodel.AppSettings{},
		Table:  "app_settings",
	},
	{
		Name:   "TaskEvent",
		Domain: &eventsdomain.TaskEvent{},
		Model:  &eventsmodel.TaskEvent{},
		Table:  "task_events",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
		},
	},
	{
		Name:   "Task",
		Domain: &taskcoredomain.Task{},
		Model:  &taskcoremodel.Task{},
		Table:  "tasks",
		ModelMigrateExtra: []any{
			&projectmodel.Project{},
		},
	},
	{
		Name:   "TaskDependency",
		Domain: &taskcoredomain.TaskDependency{},
		Model:  &taskcoremodel.TaskDependency{},
		Table:  "task_dependencies",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
		},
	},
	{
		Name:   "Project",
		Domain: &projectsdomain.Project{},
		Model:  &projectmodel.Project{},
		Table:  "projects",
	},
	{
		Name:   "ProjectContextItem",
		Domain: &projectsdomain.ProjectContextItem{},
		Model:  &projectmodel.ProjectContextItem{},
		Table:  "project_context_items",
		ModelMigrateExtra: []any{
			&projectmodel.Project{},
			&taskcoremodel.Task{},
			&cyclesmodel.TaskCycle{},
		},
	},
	{
		Name:   "ProjectContextEdge",
		Domain: &projectsdomain.ProjectContextEdge{},
		Model:  &projectmodel.ProjectContextEdge{},
		Table:  "project_context_edges",
		ModelMigrateExtra: []any{
			&projectmodel.Project{},
			&projectmodel.ProjectContextItem{},
		},
	},
	{
		Name:   "TaskContextSnapshot",
		Domain: &taskcoredomain.TaskContextSnapshot{},
		Model:  &taskcoremodel.TaskContextSnapshot{},
		Table:  "task_context_snapshots",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
			&cyclesmodel.TaskCycle{},
			&projectmodel.Project{},
		},
	},
	{
		Name:   "TaskChecklistItem",
		Domain: &checklistdomain.TaskChecklistItem{},
		Model:  &checklistmodel.TaskChecklistItem{},
		Table:  "task_checklist_items",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
		},
	},
	{
		Name:   "TaskChecklistCompletion",
		Domain: &checklistdomain.TaskChecklistCompletion{},
		Model:  &checklistmodel.TaskChecklistCompletion{},
		Table:  "task_checklist_completions",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
			&checklistmodel.TaskChecklistItem{},
		},
	},
	{
		Name:   "TaskChecklistItemCommand",
		Domain: &checklistdomain.TaskChecklistItemCommand{},
		Model:  &checklistmodel.TaskChecklistItemCommand{},
		Table:  "task_checklist_item_commands",
		ModelMigrateExtra: []any{
			&checklistmodel.TaskChecklistItem{},
		},
	},
	{
		Name:   "TaskCycle",
		Domain: &cyclesdomain.TaskCycle{},
		Model:  &cyclesmodel.TaskCycle{},
		Table:  "task_cycles",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
		},
	},
	{
		Name:   "TaskCyclePhase",
		Domain: &cyclesdomain.TaskCyclePhase{},
		Model:  &cyclesmodel.TaskCyclePhase{},
		Table:  "task_cycle_phases",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
			&cyclesmodel.TaskCycle{},
		},
	},
	{
		Name:   "TaskCycleStreamEvent",
		Domain: &cyclesdomain.TaskCycleStreamEvent{},
		Model:  &cyclesmodel.TaskCycleStreamEvent{},
		Table:  "task_cycle_stream_events",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
			&cyclesmodel.TaskCycle{},
		},
	},
	{
		Name:   "TaskCycleCriteriaReport",
		Domain: &cyclesdomain.TaskCycleCriteriaReport{},
		Model:  &cyclesmodel.TaskCycleCriteriaReport{},
		Table:  "task_cycle_criteria_reports",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
			&cyclesmodel.TaskCycle{},
			&checklistmodel.TaskChecklistItem{},
		},
	},
	{
		Name:   "TaskCycleVerifyReport",
		Domain: &cyclesdomain.TaskCycleVerifyReport{},
		Model:  &cyclesmodel.TaskCycleVerifyReport{},
		Table:  "task_cycle_verify_reports",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
			&cyclesmodel.TaskCycle{},
			&checklistmodel.TaskChecklistItem{},
		},
	},
	{
		Name:   "TaskCycleCommandRun",
		Domain: &cyclesdomain.TaskCycleCommandRun{},
		Model:  &cyclesmodel.TaskCycleCommandRun{},
		Table:  "task_cycle_command_runs",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
			&cyclesmodel.TaskCycle{},
			&checklistmodel.TaskChecklistItem{},
		},
	},
	{
		Name:   "TaskCycleCommit",
		Domain: &cyclesdomain.TaskCycleCommit{},
		Model:  &cyclesmodel.TaskCycleCommit{},
		Table:  "task_cycle_commits",
		ModelMigrateExtra: []any{
			&taskcoremodel.Task{},
			&cyclesmodel.TaskCycle{},
		},
	},
	{
		Name:   "TaskDraft",
		Domain: &composedomain.TaskDraft{},
		Model:  &composemodel.TaskDraft{},
		Table:  "task_drafts",
	},
	{
		Name:   "TaskTemplate",
		Domain: &composedomain.TaskTemplate{},
		Model:  &composemodel.TaskTemplate{},
		Table:  "task_templates",
	},
}
