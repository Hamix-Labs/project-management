package model

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/parity"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
)

// ParityPair is the BC-local name for the shared parity registry entry type.
type ParityPair = parity.Pair

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
