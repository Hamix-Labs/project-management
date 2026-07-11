// Package store implements GORM persistence for the settings bounded context.
package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/settings/store/internal/settings"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
	"gorm.io/gorm"
)

// Store is the GORM-backed persistence facade for app_settings.
type Store struct {
	db *gorm.DB
}

// NewStore returns a Store backed by db.
func NewStore(db *gorm.DB) *Store {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.store.NewStore")
	return &Store{db: db}
}

// AppSettings is the singleton runtime-settings row.
type AppSettings = domain.AppSettings

// SettingsPatch is the partial-update payload for UpdateSettings.
type SettingsPatch = contract.SettingsPatch

// GetSettings returns the singleton app_settings row, creating it from
// hard-coded defaults on first read.
func (s *Store) GetSettings(ctx context.Context) (AppSettings, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.store.GetSettings")
	return settings.Get(ctx, s.db)
}

// UpdateSettings applies a partial patch to the singleton row.
func (s *Store) UpdateSettings(ctx context.Context, patch SettingsPatch) (AppSettings, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.store.UpdateSettings")
	return settings.Update(ctx, s.db, patch)
}

// DB exposes the underlying GORM handle for tests that assert row counts.
func (s *Store) DB() *gorm.DB { return s.db }
