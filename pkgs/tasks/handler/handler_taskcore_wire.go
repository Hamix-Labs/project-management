package handler

import (
	"context"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/registry"
	settingshandler "github.com/AlexsanderHamir/Hamix/pkgs/settings/handler"
	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcorehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

// registryRunnerPorts adapts agents/runner/registry for BC handler DI
// (taskcore RunnerValidator, settings RunnerModelLister).
type registryRunnerPorts struct{}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (registryRunnerPorts) ValidateRunner(id string) error {
	_, err := registry.Lookup(id)
	return err
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (registryRunnerPorts) ListModels(ctx context.Context, runnerID, binaryPath string, timeout time.Duration) ([]settingshandler.RunnerModel, string, error) {
	models, resolved, err := registry.ListModelsForRunner(ctx, runnerID, binaryPath, timeout)
	if err != nil {
		return nil, resolved, err
	}
	out := make([]settingshandler.RunnerModel, 0, len(models))
	for _, m := range models {
		out = append(out, settingshandler.RunnerModel{ID: m.ID, Label: m.Label})
	}
	return out, resolved, nil
}

//funclogmeasure:skip category=hot-path reason="Composition wiring only; operation trace is emitted by taskcore handlers."
func (h *Handler) taskcoreDeps() taskcorehandler.Deps {
	return taskcorehandler.Deps{
		Tasks:      h.store,
		Settings:   h.store,
		GitCompose: h,
		HTTP:       taskcoreHTTP{},
		Runners:    registryRunnerPorts{},
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
