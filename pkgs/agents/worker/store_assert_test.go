package worker_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestConcreteStoreSatisfiesWorkerStore(t *testing.T) {
	var _ worker.Store = (*store.Store)(nil)
}
