package model

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/parity"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// ParityPair is the BC-local name for the shared parity registry entry type.
type ParityPair = parity.Pair

// ParityPairs is the cycle-only registry for field-parity tests in this package.
var ParityPairs = []ParityPair{
	{
		Name:   "TaskCycle",
		Domain: &cyclesdomain.TaskCycle{},
		Model:  &TaskCycle{},
		Table:  "task_cycles",
	},
	{
		Name:   "TaskCyclePhase",
		Domain: &cyclesdomain.TaskCyclePhase{},
		Model:  &TaskCyclePhase{},
		Table:  "task_cycle_phases",
		ModelMigrateExtra: []any{
			&TaskCycle{},
		},
	},
	{
		Name:   "TaskCycleStreamEvent",
		Domain: &cyclesdomain.TaskCycleStreamEvent{},
		Model:  &TaskCycleStreamEvent{},
		Table:  "task_cycle_stream_events",
		ModelMigrateExtra: []any{
			&TaskCycle{},
		},
	},
	{
		Name:   "TaskCycleCriteriaReport",
		Domain: &cyclesdomain.TaskCycleCriteriaReport{},
		Model:  &TaskCycleCriteriaReport{},
		Table:  "task_cycle_criteria_reports",
		ModelMigrateExtra: []any{
			&TaskCycle{},
			&checklistmodel.TaskChecklistItem{},
		},
	},
	{
		Name:   "TaskCycleVerifyReport",
		Domain: &cyclesdomain.TaskCycleVerifyReport{},
		Model:  &TaskCycleVerifyReport{},
		Table:  "task_cycle_verify_reports",
		ModelMigrateExtra: []any{
			&TaskCycle{},
			&checklistmodel.TaskChecklistItem{},
		},
	},
	{
		Name:   "TaskCycleCommandRun",
		Domain: &cyclesdomain.TaskCycleCommandRun{},
		Model:  &TaskCycleCommandRun{},
		Table:  "task_cycle_command_runs",
		ModelMigrateExtra: []any{
			&TaskCycle{},
			&checklistmodel.TaskChecklistItem{},
		},
	},
	{
		Name:   "TaskCycleCommit",
		Domain: &cyclesdomain.TaskCycleCommit{},
		Model:  &TaskCycleCommit{},
		Table:  "task_cycle_commits",
		ModelMigrateExtra: []any{
			&TaskCycle{},
		},
	},
}
