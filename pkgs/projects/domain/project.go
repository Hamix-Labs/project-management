package domain

import "time"

// Project is shared context memory for a long-running body of work.
//
// RepositoryID ties a project to exactly one global repository (ADR-0037); the
// repository must exist first. IsDefault marks the non-deletable system default
// seeded when a repo is registered (ADR-0042). Plain indexed nullable column
// (no FK constraint, same pattern as Task git-binding columns).
type Project struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Status         ProjectStatus `json:"status"`
	ContextSummary string        `json:"context_summary"`
	RepositoryID   *string       `json:"repository_id,omitempty"`
	IsDefault      bool          `json:"is_default"`
	// NextTaskNumber is the next tasks.number to allocate for this project.
	// Persistence counter only — omitted from the project HTTP API.
	NextTaskNumber int       `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ProjectContextItem is a human-inspectable memory item attached to a project.
type ProjectContextItem struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	Tag           string    `json:"tag"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Body          string    `json:"body"`
	SourceTaskID  *string   `json:"source_task_id,omitempty"`
	SourceCycleID *string   `json:"source_cycle_id,omitempty"`
	CreatedBy     Actor     `json:"created_by"`
	Pinned        bool      `json:"pinned"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ProjectContextEdge is a user-curated relationship between two context nodes
// owned by the same project.
type ProjectContextEdge struct {
	ID              string                 `json:"id"`
	ProjectID       string                 `json:"project_id"`
	SourceContextID string                 `json:"source_context_id"`
	TargetContextID string                 `json:"target_context_id"`
	Relation        ProjectContextRelation `json:"relation"`
	Strength        int                    `json:"strength"`
	Note            string                 `json:"note"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}
