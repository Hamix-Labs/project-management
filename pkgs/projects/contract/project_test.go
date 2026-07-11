package contract_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/projects/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestStore_implements_ProjectStore(t *testing.T) {
	var _ contract.ProjectStore = (*store.Store)(nil)
}
