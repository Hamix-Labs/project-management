package contract_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestStore_implements_TaskCRUDStore(t *testing.T) {
	var _ contract.TaskCRUDStore = (*store.Store)(nil)
}
