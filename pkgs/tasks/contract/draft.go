package contract

import (
	"context"
	"encoding/json"
)

// DraftStore covers task draft persistence.
type DraftStore interface {
	ListDrafts(ctx context.Context, limit int) ([]DraftSummary, error)
	SaveDraft(ctx context.Context, id, name string, payload json.RawMessage) (*DraftSummary, error)
	GetDraft(ctx context.Context, id string) (*DraftDetail, error)
	DeleteDraft(ctx context.Context, id string) error
}
