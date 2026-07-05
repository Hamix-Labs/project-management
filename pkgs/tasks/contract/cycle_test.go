package contract_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestStore_implements_CycleStore(t *testing.T) {
	var _ contract.CycleStore = (*store.Store)(nil)
}
