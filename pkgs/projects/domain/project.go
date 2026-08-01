package domain

import "time"

// Project is a long-lived container for tasks.
//
// User projects (IsDefault=false) require RepositoryID tying them to exactly
// one registered repository (ADR-0037). The single system Default (IsDefault,
// ADR-0094) has a null RepositoryID and may hold tasks from any repository.
// Plain indexed nullable column (no FK constraint, same pattern as Task
// git-binding columns).
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
