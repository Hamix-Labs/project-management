package contract_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
)

func TestStore_implements_GitInventoryStore(t *testing.T) {
	var _ contract.GitInventoryStore = (*store.Store)(nil)
}

func TestStore_implements_GitWriteStore(t *testing.T) {
	var _ contract.GitWriteStore = (*store.Store)(nil)
}
