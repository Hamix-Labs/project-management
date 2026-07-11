package storefake_test

import (
	"testing"

	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/storefake"
)

func TestTaskCRUDFake_implements_TaskCRUDStore(t *testing.T) {
	t.Parallel()
	var _ taskcorecontract.TaskCRUDStore = (*storefake.TaskCRUDFake)(nil)
}

func TestHandlerStoreFake_implements_HandlerStore(t *testing.T) {
	t.Parallel()
	var _ handler.HandlerStore = (*storefake.HandlerStoreFake)(nil)
}
