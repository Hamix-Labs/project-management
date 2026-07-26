package contract

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
)

// CreateProjectInput is the store input for creating a project.
type CreateProjectInput struct {
	ID             string
	Name           string
	Description    string
	ContextSummary string
	RepositoryID   *string
}

// UpdateProjectInput is a partial patch for project metadata.
type UpdateProjectInput struct {
	Name           *string
	Description    *string
	Status         *domain.ProjectStatus
	ContextSummary *string
}

// CreateProjectContextInput is the store input for appending a project context item.
type CreateProjectContextInput struct {
	ID            string
	Tag           string
	Title         string
	Description   string
	Body          string
	SourceTaskID  *string
	SourceCycleID *string
	CreatedBy     domain.Actor
	Pinned        bool
}

// UpdateProjectContextInput is a partial patch for one project context item.
type UpdateProjectContextInput struct {
	Tag         *string
	Title       *string
	Description *string
	Body        *string
	Pinned      *bool
}

// CreateProjectContextEdgeInput is the store input for connecting two context nodes.
type CreateProjectContextEdgeInput struct {
	ID              string
	SourceContextID string
	TargetContextID string
	Relation        domain.ProjectContextRelation
	Strength        int
	Note            string
}

// UpdateProjectContextEdgeInput is a partial patch for one project context edge.
type UpdateProjectContextEdgeInput struct {
	Relation *domain.ProjectContextRelation
	Strength *int
	Note     *string
}
