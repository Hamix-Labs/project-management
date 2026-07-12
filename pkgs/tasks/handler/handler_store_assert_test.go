package handler

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
)

func TestStoreImplementsHandlerAPI(t *testing.T) {
	t.Parallel()
	var _ HandlerStore = (*composition.API)(nil)
}
