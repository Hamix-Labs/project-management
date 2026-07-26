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
}
