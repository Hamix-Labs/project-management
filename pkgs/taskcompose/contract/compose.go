package contract

import (
	"context"
	"encoding/json"
)

// ListTemplatesInput is the filter/sort options for ListTemplates.
type ListTemplatesInput struct {
	Limit int
	Q     string
	Sort  string
	Order string
	Tag   string
}

// ComposeStore covers task draft and template persistence.
type ComposeStore interface {
	ListDrafts(ctx context.Context, limit int) ([]DraftSummary, error)
	SaveDraft(ctx context.Context, id, name string, payload json.RawMessage) (*DraftSummary, error)
	GetDraft(ctx context.Context, id string) (*DraftDetail, error)
	DeleteDraft(ctx context.Context, id string) error

	ListTemplates(ctx context.Context, in ListTemplatesInput) ([]TemplateSummary, error)
	SaveTemplate(ctx context.Context, id, name string, payload json.RawMessage) (*TemplateSummary, error)
	GetTemplate(ctx context.Context, id string) (*TemplateDetail, error)
	PatchTemplate(ctx context.Context, id string, name *string, payload json.RawMessage) (*TemplateDetail, error)
	DeleteTemplate(ctx context.Context, id string) error
	IncrementTemplateInstantiateCounts(ctx context.Context, counts map[string]int) error
}
