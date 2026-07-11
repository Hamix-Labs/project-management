package contract

import (
	"encoding/json"
	"time"
)

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
