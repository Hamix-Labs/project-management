package contract

import (
	"encoding/json"
	"time"

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
	Kind          domain.ProjectContextKind
	Title         string
	Body          string
	SourceTaskID  *string
	SourceCycleID *string
	CreatedBy     domain.Actor
	Pinned        bool
}

// UpdateProjectContextInput is a partial patch for one project context item.
type UpdateProjectContextInput struct {
	Kind   *domain.ProjectContextKind
	Title  *string
	Body   *string
	Pinned *bool
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

// DraftSummary is the listing-row shape for task drafts.
type DraftSummary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

// DraftDetail is the GET-by-id body shape for task drafts.
type DraftDetail struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Payload   json.RawMessage `json:"payload"`
	UpdatedAt time.Time       `json:"updated_at"`
	CreatedAt time.Time       `json:"created_at"`
}

// TemplateSummary is the listing-row shape for task templates.
type TemplateSummary struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	UpdatedAt        time.Time `json:"updated_at"`
	CreatedAt        time.Time `json:"created_at"`
	PrimaryTag       string    `json:"primary_tag,omitempty"`
	InstantiateCount int       `json:"instantiate_count"`
}

// TemplateDetail is the GET-by-id body shape for task templates.
type TemplateDetail = DraftDetail
