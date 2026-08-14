package domain

import "time"

// FormSnapshot is the create-task form state the operator has entered so far.
// The agent reads it for context; only Prompt is writable via MCP in v1.
type FormSnapshot struct {
	Title       string   `json:"title,omitempty"`
	Prompt      string   `json:"prompt,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
	Criteria    []string `json:"criteria,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	CursorModel string   `json:"cursor_model,omitempty"`

	// UpdatedAt is set by the store; callers ignore it on write.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Session is the in-memory record for one draft-assist modal.
type Session struct {
	ID         string
	Nonce      string
	WorktreeID string
	Snapshot   FormSnapshot
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
