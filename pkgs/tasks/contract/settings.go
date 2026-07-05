package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// SettingsStore covers singleton app_settings read/write.
type SettingsStore interface {
	GetSettings(ctx context.Context) (domain.AppSettings, error)
	UpdateSettings(ctx context.Context, patch SettingsPatch) (domain.AppSettings, error)
}
