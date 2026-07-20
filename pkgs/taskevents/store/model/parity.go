package model

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/parity"
	eventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// ParityPair is the BC-local name for the shared parity registry entry type.
type ParityPair = parity.Pair

// ParityPairs is the taskevents registry for field-parity tests in this package.
var ParityPairs = []ParityPair{
	{
		Name:   "TaskEvent",
		Domain: &eventsdomain.TaskEvent{},
		Model:  &TaskEvent{},
		Table:  "task_events",
		ModelMigrateExtra: []any{
			&TaskRow{},
		},
	},
}
