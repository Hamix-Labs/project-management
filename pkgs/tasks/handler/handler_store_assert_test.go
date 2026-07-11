package handler

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestStoreImplementsHandlerAPI(t *testing.T) {
	t.Parallel()
	var _ HandlerStore = (*store.Store)(nil)
}
