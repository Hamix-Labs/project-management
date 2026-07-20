package model

import (
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/parity"
)

// ParityPair is the BC-local name for the shared parity registry entry type.
type ParityPair = parity.Pair

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
