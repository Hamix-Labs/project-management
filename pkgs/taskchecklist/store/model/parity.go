package model

import (
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
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

// ParityPairs is the checklist registry for field-parity tests in this package.
var ParityPairs = []ParityPair{
	{
		Name:   "TaskChecklistItem",
		Domain: &checklistdomain.TaskChecklistItem{},
		Model:  &TaskChecklistItem{},
		Table:  "task_checklist_items",
		ModelMigrateExtra: []any{
			&TaskRow{},
		},
	},
	{
		Name:   "TaskChecklistCompletion",
		Domain: &checklistdomain.TaskChecklistCompletion{},
		Model:  &TaskChecklistCompletion{},
		Table:  "task_checklist_completions",
		ModelMigrateExtra: []any{
			&TaskRow{},
			&TaskChecklistItem{},
		},
	},
	{
		Name:   "TaskChecklistItemCommand",
		Domain: &checklistdomain.TaskChecklistItemCommand{},
		Model:  &TaskChecklistItemCommand{},
		Table:  "task_checklist_item_commands",
		ModelMigrateExtra: []any{
			&TaskChecklistItem{},
		},
	},
}
