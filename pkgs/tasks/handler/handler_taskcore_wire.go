package handler

import (
	taskcorehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

//funclogmeasure:skip category=hot-path reason="Composition wiring only; operation trace is emitted by taskcore handlers."
func (h *Handler) taskcoreDeps() taskcorehandler.Deps {
	return taskcorehandler.Deps{
		Tasks:      h.store,
		Settings:   h.store,
		GitCompose: h,
		NotifyChange: func(typ realtime.ChangeType, id string) {
			h.notifyChange(typ, id)
		},
		NotifyTaskChanged: func(typ realtime.ChangeType, id string, data any) {
			h.notifyTaskChanged(typ, id, data)
		},
	}
}

//funclogmeasure:skip category=hot-path reason="Composition wiring only; operation trace is emitted by taskcore handlers."
func (h *Handler) taskcoreHandler() *taskcorehandler.Handler {
	return taskcorehandler.New(h.taskcoreDeps())
}
