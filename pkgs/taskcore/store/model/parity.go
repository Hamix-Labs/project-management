package model

import (
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// ParityPairs is the taskcore registry for field-parity tests in this package.
var ParityPairs = []ParityPair{
	{
		Name:   "Task",
		Domain: &taskcoredomain.Task{},
		Model:  &Task{},
		Table:  "tasks",
		ModelMigrateExtra: []any{
			&ProjectRow{},
		},
	},
	{
		Name:   "TaskDependency",
		Domain: &taskcoredomain.TaskDependency{},
		Model:  &TaskDependency{},
		Table:  "task_dependencies",
		ModelMigrateExtra: []any{
			&Task{},
		},
	},
	{
		Name:   "TaskContextSnapshot",
		Domain: &taskcoredomain.TaskContextSnapshot{},
		Model:  &TaskContextSnapshot{},
		Table:  "task_context_snapshots",
		ModelMigrateExtra: []any{
			&Task{},
			&CycleRow{},
			&ProjectRow{},
		},
	},
}

// ParityPair binds a domain struct prototype to its model counterpart for
// schema- and field-parity guards.
type ParityPair struct {
	Name              string
	Domain            any
	Model             any
	Table             string
	ModelMigrateExtra []any
}
