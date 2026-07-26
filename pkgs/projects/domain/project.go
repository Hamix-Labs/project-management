package domain

import "time"

// Project is a long-lived container for tasks bound to one repository.
//
// RepositoryID ties a project to exactly one global repository (ADR-0037); the
// repository must exist first. IsDefault marks the non-deletable system default
// seeded when a repo is registered (ADR-0042). Plain indexed nullable column
// (no FK constraint, same pattern as Task git-binding columns).
type Project struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Status       ProjectStatus `json:"status"`
	RepositoryID *string       `json:"repository_id,omitempty"`
	IsDefault    bool          `json:"is_default"`
	// NextTaskNumber is the next tasks.number to allocate for this project.
	// Persistence counter only — omitted from the project HTTP API.
	NextTaskNumber int       `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
