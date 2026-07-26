package model

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/parity"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// ParityPair is the BC-local name for the shared parity registry entry type.
type ParityPair = parity.Pair

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
}
