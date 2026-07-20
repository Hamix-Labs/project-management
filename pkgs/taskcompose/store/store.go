// Package store implements GORM persistence for task drafts and templates.
package store

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/store/internal/drafts"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/store/internal/templates"
	"gorm.io/gorm"
)

// Store is the GORM-backed persistence facade for task drafts and templates.
type Store struct {
	db *gorm.DB
}

// NewStore returns a Store backed by db.
func NewStore(db *gorm.DB) *Store {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.store.NewStore")
	return &Store{db: db}
}

type (
	// DraftSummary is the listing-row shape for task drafts.
	DraftSummary = drafts.Summary
	// DraftDetail is the GET-by-id body shape for task drafts.
	DraftDetail = drafts.Detail
	// TemplateSummary is the listing-row shape for task templates.
	TemplateSummary = templates.Summary
	// TemplateDetail is the GET-by-id body shape for task templates.
	TemplateDetail = templates.Detail
)

// DeleteDraftByIDInTx removes a draft row inside an open transaction (Create-from-draft).
//
//funclogmeasure:skip category=delegate-already-logs reason="Package-level forwarder; drafts.DeleteByIDInTx emits trace at the store chokepoint."
func DeleteDraftByIDInTx(tx *gorm.DB, id string) error {
	return drafts.DeleteByIDInTx(tx, id)
}

func (s *Store) SaveDraft(ctx context.Context, id, name string, payload json.RawMessage) (*DraftSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.store.SaveDraft")
	return drafts.Save(ctx, s.db, id, name, payload)
}

func (s *Store) ListDrafts(ctx context.Context, limit int) ([]DraftSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.store.ListDrafts")
	return drafts.List(ctx, s.db, limit)
}

func (s *Store) GetDraft(ctx context.Context, id string) (*DraftDetail, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.store.GetDraft")
	return drafts.Get(ctx, s.db, id)
}

func (s *Store) DeleteDraft(ctx context.Context, id string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.store.DeleteDraft")
	return drafts.Delete(ctx, s.db, id)
}

func (s *Store) SaveTemplate(ctx context.Context, id, name string, payload json.RawMessage) (*TemplateSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.store.SaveTemplate")
	return templates.Save(ctx, s.db, id, name, payload)
}

func (s *Store) ListTemplates(ctx context.Context, limit int, q, sort, order, tag string) ([]TemplateSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.store.ListTemplates")
	return templates.List(ctx, s.db, limit, q, sort, order, tag)
}

func (s *Store) IncrementTemplateInstantiateCounts(ctx context.Context, counts map[string]int) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.store.IncrementTemplateInstantiateCounts")
	return templates.IncrementInstantiateCounts(ctx, s.db, counts)
}

func (s *Store) GetTemplate(ctx context.Context, id string) (*TemplateDetail, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.store.GetTemplate")
	return templates.Get(ctx, s.db, id)
}

func (s *Store) PatchTemplate(ctx context.Context, id string, name *string, payload json.RawMessage) (*TemplateDetail, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.store.PatchTemplate")
	return templates.Patch(ctx, s.db, id, name, payload)
}

func (s *Store) DeleteTemplate(ctx context.Context, id string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.store.DeleteTemplate")
	return templates.Delete(ctx, s.db, id)
}
