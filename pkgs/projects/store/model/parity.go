package model

import (
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
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

// ParityPairs is the projects registry for field-parity tests in this package.
var ParityPairs = []ParityPair{
	{
		Name:   "Project",
		Domain: &projectsdomain.Project{},
		Model:  &Project{},
		Table:  "projects",
	},
	{
		Name:   "ProjectContextItem",
		Domain: &projectsdomain.ProjectContextItem{},
		Model:  &ProjectContextItem{},
		Table:  "project_context_items",
		ModelMigrateExtra: []any{
			&Project{},
			&TaskRow{},
			&CycleRow{},
		},
	},
	{
		Name:   "ProjectContextEdge",
		Domain: &projectsdomain.ProjectContextEdge{},
		Model:  &ProjectContextEdge{},
		Table:  "project_context_edges",
		ModelMigrateExtra: []any{
			&Project{},
			&ProjectContextItem{},
		},
	},
}
