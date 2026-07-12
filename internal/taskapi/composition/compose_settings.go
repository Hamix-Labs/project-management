package composition

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
)

// GetSettings returns the singleton app_settings row, creating it from
// hard-coded defaults on first read so callers always see a populated
// value. There is no env-var fallback — the DB row is the only source
// of truth.
func (a *API) GetSettings(ctx context.Context) (settingsdomain.AppSettings, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetSettings")
	return a.settings.GetSettings(ctx)
}

// UpdateSettings applies a partial patch to the singleton row inside a
// transaction. Returns the post-update value so the caller (the HTTP
// handler) can echo the canonical row in the PATCH response without an
// extra round-trip.
func (a *API) UpdateSettings(ctx context.Context, patch settingscontract.SettingsPatch) (settingsdomain.AppSettings, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateSettings")
	return a.settings.UpdateSettings(ctx, patch)
}
