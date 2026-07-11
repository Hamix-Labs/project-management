package contract_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestStore_implements_SettingsStore(t *testing.T) {
	var _ contract.SettingsStore = (*store.Store)(nil)
}
