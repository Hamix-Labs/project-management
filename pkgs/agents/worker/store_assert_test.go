package worker_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
)

func TestConcreteStoreSatisfiesWorkerStore(t *testing.T) {
	var _ worker.Store = (*composition.API)(nil)
}
