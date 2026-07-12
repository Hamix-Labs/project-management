package model

import composedomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/domain"

// ParityPair binds a domain struct prototype to its model counterpart for
// schema- and field-parity guards.
type ParityPair struct {
	Name              string
	Domain            any
	Model             any
	Table             string
	ModelMigrateExtra []any
}

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
