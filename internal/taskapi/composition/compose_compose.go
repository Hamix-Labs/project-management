package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"log/slog"

	composecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/contract"
)

func (a *API) SaveDraft(ctx context.Context, id, name string, payload json.RawMessage) (*composecontract.DraftSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SaveDraft")
	return a.compose.SaveDraft(ctx, id, name, payload)
}

func (a *API) ListDrafts(ctx context.Context, limit int) ([]composecontract.DraftSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListDrafts")
	return a.compose.ListDrafts(ctx, limit)
}

func (a *API) GetDraft(ctx context.Context, id string) (*composecontract.DraftDetail, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetDraft")
	return a.compose.GetDraft(ctx, id)
}

func (a *API) DeleteDraft(ctx context.Context, id string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteDraft")
	return a.compose.DeleteDraft(ctx, id)
}

func (a *API) SaveTemplate(ctx context.Context, id, name string, payload json.RawMessage) (*composecontract.TemplateSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SaveTemplate")
	return a.compose.SaveTemplate(ctx, id, name, payload)
}

func (a *API) ListTemplates(ctx context.Context, limit int, q, sort, order, tag string) ([]composecontract.TemplateSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListTemplates")
	return a.compose.ListTemplates(ctx, limit, q, sort, order, tag)
}

func (a *API) IncrementTemplateInstantiateCounts(ctx context.Context, counts map[string]int) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.IncrementTemplateInstantiateCounts")
	return a.compose.IncrementTemplateInstantiateCounts(ctx, counts)
}

func (a *API) GetTemplate(ctx context.Context, id string) (*composecontract.TemplateDetail, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetTemplate")
	return a.compose.GetTemplate(ctx, id)
}

func (a *API) PatchTemplate(ctx context.Context, id string, name *string, payload json.RawMessage) (*composecontract.TemplateDetail, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.PatchTemplate")
	return a.compose.PatchTemplate(ctx, id, name, payload)
}

func (a *API) DeleteTemplate(ctx context.Context, id string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteTemplate")
	return a.compose.DeleteTemplate(ctx, id)
}
