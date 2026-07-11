package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
)

// AppSettings is the singleton runtime-settings row returned by
// GetSettings / UpdateSettings. See pkgs/settings/domain.AppSettings for
// field semantics.
type AppSettings = settingsdomain.AppSettings

// SettingsPatch is the partial-update payload for UpdateSettings.
type SettingsPatch = contract.SettingsPatch

// GetSettings returns the singleton app_settings row, creating it from
// hard-coded defaults on first read so callers always see a populated
// value. There is no env-var fallback — the DB row is the only source
// of truth.
func (s *Store) GetSettings(ctx context.Context) (AppSettings, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetSettings")
	return s.settings.GetSettings(ctx)
}

// UpdateSettings applies a partial patch to the singleton row inside a
// transaction. Returns the post-update value so the caller (the HTTP
// handler) can echo the canonical row in the PATCH response without an
// extra round-trip.
func (s *Store) UpdateSettings(ctx context.Context, patch SettingsPatch) (AppSettings, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UpdateSettings")
	return s.settings.UpdateSettings(ctx, patch)
}
