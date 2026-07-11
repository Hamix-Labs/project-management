package model

import (
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
)

// ParityPair binds a domain struct prototype to its model counterpart for
// schema- and field-parity guards.
type ParityPair struct {
	Name   string
	Domain any
	Model  any
	Table  string
}

// ParityPairs is the single registry both parity tests iterate.
var ParityPairs = []ParityPair{
	{
		Name:   "AppSettings",
		Domain: &settingsdomain.AppSettings{},
		Model:  &AppSettings{},
		Table:  "app_settings",
	},
}
