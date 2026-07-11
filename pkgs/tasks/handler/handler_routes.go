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
	taskcycleshandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/handler"
	eventhandler "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func (h *Handler) registerRoutes(m *http.ServeMux) {
	h.registerHealthRoutes(m)
	projecthandler.Register(m, projecthandler.Deps{
		Store: h.store,
		Notify: func(typ realtime.ChangeType, id string) {
			h.notifyChange(TaskChangeType(typ), id)
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
			h.notifyScopelessChange(TaskChangeType(typ))
		},
	})
	composehandler.Register(m, composehandler.Deps{
		Compose: h.store,
		NormalizeCompose: func(ctx context.Context, raw json.RawMessage) (composehandler.NormalizeComposeResult, error) {
			payloadRaw, compose, err := h.normalizeComposePayloadRaw(ctx, raw)
			if err != nil {
				return composehandler.NormalizeComposeResult{}, err
			}
			return composehandler.NormalizeComposeResult{Payload: payloadRaw, Title: compose.Title}, nil
		},
		InstantiateFromTemplate: func(ctx context.Context, r *http.Request, op string, payload json.RawMessage, by domain.Actor) (*domain.Task, error) {
			compose, err := decodeComposePayload(payload)
			if err != nil {
				return nil, err
			}
			return h.createTaskFromComposeJSON(ctx, r, op, compose, createTaskComposeOpts{
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
	h.registerTaskRoutes(m)
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
func (h *Handler) registerTaskRoutes(m *http.ServeMux) {
	m.Handle("POST /tasks", http.HandlerFunc(h.create))
	m.Handle("GET /tasks", http.HandlerFunc(h.list))
	m.Handle("GET /tasks/stats", http.HandlerFunc(h.stats))
	m.Handle("GET /tasks/{id}/dependencies", http.HandlerFunc(h.listTaskDependencies))
	m.Handle("POST /tasks/{id}/dependencies", http.HandlerFunc(h.addTaskDependency))
	m.Handle("DELETE /tasks/{id}/dependencies/{depId}", http.HandlerFunc(h.removeTaskDependency))
	m.Handle("PATCH /tasks/{id}/gate", http.HandlerFunc(h.patchTaskGate))
	m.Handle("POST /tasks/{id}/retry", http.HandlerFunc(h.postTaskRetry))
	m.Handle("GET /tasks/{id}", http.HandlerFunc(h.get))
	m.Handle("PATCH /tasks/{id}", http.HandlerFunc(h.patch))
	m.Handle("DELETE /tasks/{id}", http.HandlerFunc(h.delete))
}

//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func (h *Handler) registerMiscRoutes(m *http.ServeMux) {
	// /v1/rum is the SPA-side Real User Monitoring beacon. Documented
	// in docs/architecture.md; the browser ships batches via
	// `navigator.sendBeacon` so the server returns 204 with no body.
	// Rate-limited via the global per-IP middleware (WithRateLimit),
	// not separately, so a misbehaving SPA cannot amplify a load
	// incident into a metrics-storage bill.
	m.Handle("POST /v1/rum", http.HandlerFunc(h.postRUM))
	// /v1/bootstrap is the cold-start aggregate the SPA uses to seed
	// its TanStack Query cache from a single round trip — combines
	// settings, root tasks page, stats, projects, and drafts head.
	// Documented in docs/api.md; clients must tolerate 5xx and fall
	// back to per-endpoint fan-out.
	m.Handle("GET /v1/bootstrap", http.HandlerFunc(h.bootstrap))
}
