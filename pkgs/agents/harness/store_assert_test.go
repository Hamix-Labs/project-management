package harness

import (
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
)

var _ Store = (*composition.API)(nil)
