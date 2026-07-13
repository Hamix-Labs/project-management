package storefake

import (
	"context"

	settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
)

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetSettings(context.Context) (settingsdomain.AppSettings, error) {
	return settingsdomain.AppSettings{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateSettings(context.Context, settingscontract.SettingsPatch) (settingsdomain.AppSettings, error) {
	return settingsdomain.AppSettings{}, errNotImplemented
}
