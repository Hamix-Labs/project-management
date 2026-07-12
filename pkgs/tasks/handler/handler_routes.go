package handler

import (
	"context"
	"encoding/json"
	"net/http"

	gitinventoryhandler "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/handler"
	projecthandler "github.com/AlexsanderHamir/Hamix/pkgs/projects/handler"
	repohandler "github.com/AlexsanderHamir/Hamix/pkgs/repo/handler"
	runnershandler "github.com/AlexsanderHamir/Hamix/pkgs/runners/handler"
	settingshandler "github.com/AlexsanderHamir/Hamix/pkgs/settings/handler"
	checklisthandler "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/handler"
	composehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/handler"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
	taskcycleshandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/handler"
	eventhandler "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func (h *Handler) registerRoutes(m *http.ServeMux) {
	h.registerHealthRoutes(m)
	projecthandler.Register(m, projecthandler.Deps{
		Store: h.store,
		Notify: func(typ realtime.ChangeType, id string) {
			h.notifyChange(realtime.ChangeType(typ), id)
		},
	})
	gitinventoryhandler.Register(m, gitinventoryhandler.Deps{
		Read:       h.store,
		Write:      h.store,
		Projects:   h.store,
		GitService: h.git,
		HostPaths:  h.pathMap,
	})
	settingshandler.Register(m, settingshandler.Deps{
		Settings: h.store,
		GitRead:  h.store,
		Agent:    h.agent,
		Git:      h.git,
		Notify: func(typ realtime.ChangeType) {
			h.notifyScopelessChange(realtime.ChangeType(typ))
		},
	})
	tc := h.taskcoreHandler()
	composehandler.Register(m, composehandler.Deps{
		Compose: h.store,
		NormalizeCompose: func(ctx context.Context, raw json.RawMessage) (composehandler.NormalizeComposeResult, error) {
			payloadRaw, compose, err := h.normalizeComposePayloadRaw(ctx, raw)
			if err != nil {
				return composehandler.NormalizeComposeResult{}, err
			}
			return composehandler.NormalizeComposeResult{Payload: payloadRaw, Title: compose.Title}, nil
		},
		InstantiateFromTemplate: func(ctx context.Context, r *http.Request, op string, payload json.RawMessage, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
			compose, err := taskcorehandler.DecodeComposePayload(payload)
			if err != nil {
				return nil, err
			}
			return tc.CreateTaskFromComposeJSON(ctx, r, op, compose, taskcorehandler.CreateTaskComposeOpts{
				StripDependsOn:          true,
				OmitPastPickupNotBefore: true,
				InstantiateFromTemplate: true,
			}, by)
		},
	})
	checklisthandler.Register(m, checklisthandler.Deps{
		Checklist: h.store,
		NotifyTaskUpdated: func(ctx context.Context, taskID string) error {
			return h.notifyTaskUpdatedEnriched(ctx, taskID)
		},
	})
	eventhandler.Register(m, eventhandler.Deps{
		Events: h.store,
		Tasks:  h.store,
		NotifyTaskEventChanged: func(taskID string, eventSeq int64) {
			h.notifyTaskEventChanged(taskID, eventSeq)
		},
	})
	taskcycleshandler.Register(m, taskcycleshandler.Deps{
		Cycles:        h.store,
		Tasks:         h.store,
		CycleFailures: h.store,
		NotifyCycleChanged: func(ctx context.Context, taskID, cycleID string, data any) {
			if data != nil {
				h.notifyCycleChanged(taskID, cycleID, data)
				return
			}
			h.notifyCycleChange(taskID, cycleID)
		},
	})
	taskcorehandler.Register(m, h.taskcoreDeps())
	repohandler.Register(m, repohandler.Deps{Provider: h.repoProv})
	runnershandler.Register(m, runnershandler.Deps{Settings: h.store})
	h.registerMiscRoutes(m)
}

//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func (h *Handler) registerHealthRoutes(m *http.ServeMux) {
	m.Handle("GET /health", http.HandlerFunc(health))
	m.Handle("GET /health/live", http.HandlerFunc(healthLive))
	m.Handle("GET /health/ready", http.HandlerFunc(h.healthReady))
	m.Handle("GET /system/health", http.HandlerFunc(h.systemHealth))
	m.Handle("GET /events", http.HandlerFunc(h.streamEvents))
}

//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func (h *Handler) registerMiscRoutes(m *http.ServeMux) {
	m.Handle("POST /v1/rum", http.HandlerFunc(h.postRUM))
	m.Handle("GET /v1/bootstrap", http.HandlerFunc(h.bootstrap))
}
