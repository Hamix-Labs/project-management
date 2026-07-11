package contract

import (
	"context"
	"encoding/json"

	composecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/contract"
)

type (
	// DraftSummary is the listing-row shape for task drafts.
	DraftSummary = composecontract.DraftSummary
	// DraftDetail is the GET-by-id body shape for task drafts.
	DraftDetail = composecontract.DraftDetail
)

// DraftStore covers task draft persistence.
type DraftStore interface {
	ListDrafts(ctx context.Context, limit int) ([]DraftSummary, error)
	SaveDraft(ctx context.Context, id, name string, payload json.RawMessage) (*DraftSummary, error)
	GetDraft(ctx context.Context, id string) (*DraftDetail, error)
	DeleteDraft(ctx context.Context, id string) error
}
