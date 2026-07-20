package model

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/parity"
	composedomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/domain"
)

// ParityPair is the BC-local name for the shared parity registry entry type.
type ParityPair = parity.Pair

// ParityPairs is the compose registry for field-parity tests in this package.
var ParityPairs = []ParityPair{
	{
		Name:   "TaskDraft",
		Domain: &composedomain.TaskDraft{},
		Model:  &TaskDraft{},
		Table:  "task_drafts",
	},
	{
		Name:   "TaskTemplate",
		Domain: &composedomain.TaskTemplate{},
		Model:  &TaskTemplate{},
		Table:  "task_templates",
	},
}
