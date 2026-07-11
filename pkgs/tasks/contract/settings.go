package contract

import (
	"context"

	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
)

// SettingsStore covers singleton app_settings read/write.
type SettingsStore interface {
	GetSettings(ctx context.Context) (settingsdomain.AppSettings, error)
	UpdateSettings(ctx context.Context, patch SettingsPatch) (settingsdomain.AppSettings, error)
}
