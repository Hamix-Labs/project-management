package domain

import "time"

// LegacyGlobalDefaultProjectID is the pre-ADR-0042 global default project row.
// Referenced by postgres migrations for historical remaps.
const LegacyGlobalDefaultProjectID = "00000000-0000-4000-8000-000000000001"

// GlobalDefaultProjectID is the stable id for the single system Default project
// (ADR-0094). Distinct from LegacyGlobalDefaultProjectID so rev-6 cleanup that
// deleted the legacy row does not collide with the ADR-0094 seed.
const GlobalDefaultProjectID = "00000000-0000-4000-8000-0000000000df"

// DefaultProjectName is the display name for the system Default project.
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

// GlobalDefaultProject returns the ADR-0094 system Default row shape.
//
//funclogmeasure:skip category=hot-path reason="Test/migration fixture only."
func GlobalDefaultProject(now time.Time) Project {
	now = now.UTC()
	return Project{
		ID:             GlobalDefaultProjectID,
		Name:           DefaultProjectName,
		Description:    "Built-in project for tasks not assigned to a custom project.",
		Status:         ProjectStatusActive,
		IsDefault:      true,
		NextTaskNumber: 1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
