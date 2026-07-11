package domain

import (
	"encoding/json"
	"time"
)

type TaskEvent struct {
	TaskID string          `json:"task_id"`
	Seq    int64           `json:"seq"`
	At     time.Time       `json:"at"`
	Type   EventType       `json:"type"`
	By     Actor           `json:"by"`
	Data   json.RawMessage `json:"data"`

	// UserResponse is optional human-supplied text for event types that accept input (see EventTypeAcceptsUserResponse).
	UserResponse *string `json:"user_response,omitempty"`
	// UserResponseAt is set when UserResponse is written or updated (UTC).
	UserResponseAt *time.Time `json:"user_response_at,omitempty"`
	// ResponseThread is an ordered JSON array of ResponseThreadEntry (user ↔ agent messages).
	ResponseThread json.RawMessage `json:"response_thread,omitempty"`
}
