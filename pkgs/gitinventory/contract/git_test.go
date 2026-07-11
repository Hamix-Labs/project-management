package contract_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestStore_implements_GitReadStore(t *testing.T) {
	var _ contract.GitReadStore = (*store.Store)(nil)
}

func TestStore_implements_GitWriteStore(t *testing.T) {
	var _ contract.GitWriteStore = (*store.Store)(nil)
}
