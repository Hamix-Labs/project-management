package domain

import "time"

// LegacyGlobalDefaultProjectID is the pre-ADR-0042 global default project row.
// Referenced only by postgres migrations until the row is deleted.
const LegacyGlobalDefaultProjectID = "00000000-0000-4000-8000-000000000001"

// DefaultProjectName is the display name for system default projects (one per repo).
const DefaultProjectName = "Default"

// LegacyGlobalDefaultProject returns the pre-ADR-0042 global default row shape for migration tests.
//
//funclogmeasure:skip category=hot-path reason="Test/migration fixture only."
func LegacyGlobalDefaultProject(now time.Time) Project {
	return Project{
		ID:          LegacyGlobalDefaultProjectID,
		Name:        "Default project",
		Description: "Built-in project for general task context.",
		Status:      ProjectStatusActive,
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC(),
	}
}
