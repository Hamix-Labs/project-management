package domain

import (
	"encoding/json"
	"time"
)

// TaskDraft stores a resumable create-task draft payload.
type TaskDraft struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	PayloadJSON json.RawMessage `json:"payload_json"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// TaskTemplate stores a reusable task compose blueprint (not a runnable task).
type TaskTemplate struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	PayloadJSON json.RawMessage `json:"payload_json"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
