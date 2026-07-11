package contract_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store"
)

func TestStore_implements_TaskEventStore(t *testing.T) {
	var _ contract.TaskEventStore = (*store.Store)(nil)
}
