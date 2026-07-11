package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"encoding/json"
	"log/slog"

	composecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/contract"
)

type (
	// DraftSummary is the listing-row shape for task drafts.
	DraftSummary = composecontract.DraftSummary
	// DraftDetail is the GET-by-id body shape for task drafts.
	DraftDetail = composecontract.DraftDetail
	// TemplateSummary is the listing-row shape for task templates.
	TemplateSummary = composecontract.TemplateSummary
	// TemplateDetail is the GET-by-id body shape for task templates.
	TemplateDetail = composecontract.TemplateDetail
)

func (s *Store) SaveDraft(ctx context.Context, id, name string, payload json.RawMessage) (*DraftSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SaveDraft")
	return s.compose.SaveDraft(ctx, id, name, payload)
}

func (s *Store) ListDrafts(ctx context.Context, limit int) ([]DraftSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListDrafts")
	return s.compose.ListDrafts(ctx, limit)
}

func (s *Store) GetDraft(ctx context.Context, id string) (*DraftDetail, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetDraft")
	return s.compose.GetDraft(ctx, id)
}

func (s *Store) DeleteDraft(ctx context.Context, id string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteDraft")
	return s.compose.DeleteDraft(ctx, id)
}

func (s *Store) SaveTemplate(ctx context.Context, id, name string, payload json.RawMessage) (*TemplateSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SaveTemplate")
	return s.compose.SaveTemplate(ctx, id, name, payload)
}

func (s *Store) ListTemplates(ctx context.Context, limit int, q, sort, order, tag string) ([]TemplateSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListTemplates")
	return s.compose.ListTemplates(ctx, limit, q, sort, order, tag)
}

func (s *Store) IncrementTemplateInstantiateCounts(ctx context.Context, counts map[string]int) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.IncrementTemplateInstantiateCounts")
	return s.compose.IncrementTemplateInstantiateCounts(ctx, counts)
}

func (s *Store) GetTemplate(ctx context.Context, id string) (*TemplateDetail, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetTemplate")
	return s.compose.GetTemplate(ctx, id)
}

func (s *Store) PatchTemplate(ctx context.Context, id string, name *string, payload json.RawMessage) (*TemplateDetail, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.PatchTemplate")
	return s.compose.PatchTemplate(ctx, id, name, payload)
}

func (s *Store) DeleteTemplate(ctx context.Context, id string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteTemplate")
	return s.compose.DeleteTemplate(ctx, id)
}
