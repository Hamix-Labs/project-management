package contract_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store"
)

func TestStore_implements_TaskEventStore(t *testing.T) {
	var _ contract.TaskEventStore = (*store.Store)(nil)
}

func TestStore_implements_TaskActivityStore(t *testing.T) {
	var _ contract.TaskActivityStore = (*store.Store)(nil)
}
