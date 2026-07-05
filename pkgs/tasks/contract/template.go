package contract

import (
	"context"
	"encoding/json"
)

// TemplateStore covers task template persistence.
type TemplateStore interface {
	ListTemplates(ctx context.Context, limit int, q, sort, order, tag string) ([]TemplateSummary, error)
	SaveTemplate(ctx context.Context, id, name string, payload json.RawMessage) (*TemplateSummary, error)
	GetTemplate(ctx context.Context, id string) (*TemplateDetail, error)
	PatchTemplate(ctx context.Context, id string, name *string, payload json.RawMessage) (*TemplateDetail, error)
	DeleteTemplate(ctx context.Context, id string) error
	IncrementTemplateInstantiateCounts(ctx context.Context, counts map[string]int) error
}
