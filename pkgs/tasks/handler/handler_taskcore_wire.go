package handler

import (
	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcorehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

//funclogmeasure:skip category=hot-path reason="Composition wiring only; operation trace is emitted by taskcore handlers."
func (h *Handler) taskcoreDeps() taskcorehandler.Deps {
	return taskcorehandler.Deps{
		Tasks:      h.store,
		Settings:   h.store,
		GitCompose: h,
		HTTP:       taskcoreHTTP{},
		NotifyChange: func(typ taskcorecontract.ChangeType, id string) {
			h.notifyChange(realtime.ChangeType(typ), id)
		},
		NotifyTaskChanged: func(typ taskcorecontract.ChangeType, id string, data any) {
			h.notifyTaskChanged(realtime.ChangeType(typ), id, data)
		},
	}
}

//funclogmeasure:skip category=hot-path reason="Composition wiring only; operation trace is emitted by taskcore handlers."
func (h *Handler) taskcoreHandler() *taskcorehandler.Handler {
	return taskcorehandler.New(h.taskcoreDeps())
}
