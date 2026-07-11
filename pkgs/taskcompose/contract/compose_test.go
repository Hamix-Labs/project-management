package contract_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/store"
)

func TestStore_implements_ComposeStore(t *testing.T) {
	var _ contract.ComposeStore = (*store.Store)(nil)
}
