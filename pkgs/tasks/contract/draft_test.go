package contract_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestStore_implements_DraftStore(t *testing.T) {
	var _ contract.DraftStore = (*store.Store)(nil)
}
