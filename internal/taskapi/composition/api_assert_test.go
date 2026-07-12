package composition

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/worker"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/wire"
)

func TestAPIImplementsHandlerAPI(t *testing.T) {
	t.Parallel()
	var _ wire.HandlerAPI = (*API)(nil)
}

func TestAPIImplementsWorkerStore(t *testing.T) {
	t.Parallel()
	var _ worker.Store = (*API)(nil)
}
