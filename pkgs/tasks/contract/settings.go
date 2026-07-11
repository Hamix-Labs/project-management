package contract

import settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"

// SettingsStore covers singleton app_settings read/write.
type SettingsStore = settingscontract.SettingsStore

// SettingsPatch is the partial-update payload for app_settings.
type SettingsPatch = settingscontract.SettingsPatch
