package model

import (
	eventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// ParityPair binds a domain struct prototype to its model counterpart for
// schema- and field-parity guards.
type ParityPair struct {
	Name              string
	Domain            any
	Model             any
	Table             string
	ModelMigrateExtra []any
}

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
