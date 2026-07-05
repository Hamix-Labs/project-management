package storefake_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/storefake"
)

func TestTaskCRUDFake_implements_TaskCRUDStore(t *testing.T) {
	t.Parallel()
	var _ contract.TaskCRUDStore = (*storefake.TaskCRUDFake)(nil)
}

func TestHandlerStoreFake_implements_HandlerStore(t *testing.T) {
	t.Parallel()
	var _ contract.HandlerStore = (*storefake.HandlerStoreFake)(nil)
}
