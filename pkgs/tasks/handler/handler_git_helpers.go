package handler

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) gitService() gitwork.Service {
	if h.git != nil {
		return h.git
	}
	return gitwork.New()
}
