package handler

import (
	"net/http"

	gitinventoryhandler "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/handler"
	projecthandler "github.com/AlexsanderHamir/Hamix/pkgs/projects/handler"
	settingshandler "github.com/AlexsanderHamir/Hamix/pkgs/settings/handler"
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
	h.registerTaskDraftTemplateRoutes(m)
	h.registerTaskRoutes(m)
	h.registerRepoRoutes(m)
	h.registerRunnerRoutes(m)
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
func (h *Handler) registerTaskDraftTemplateRoutes(m *http.ServeMux) {
	m.Handle("GET /task-drafts", http.HandlerFunc(h.listTaskDrafts))
	m.Handle("POST /task-drafts", http.HandlerFunc(h.saveTaskDraft))
	m.Handle("GET /task-drafts/{id}", http.HandlerFunc(h.getTaskDraft))
	m.Handle("DELETE /task-drafts/{id}", http.HandlerFunc(h.deleteTaskDraft))
	m.Handle("GET /task-templates", http.HandlerFunc(h.listTaskTemplates))
	m.Handle("POST /task-templates", http.HandlerFunc(h.saveTaskTemplate))
	m.Handle("GET /task-templates/{id}", http.HandlerFunc(h.getTaskTemplate))
	m.Handle("PATCH /task-templates/{id}", http.HandlerFunc(h.patchTaskTemplate))
	m.Handle("DELETE /task-templates/{id}", http.HandlerFunc(h.deleteTaskTemplate))
	m.Handle("POST /task-templates/instantiate", http.HandlerFunc(h.instantiateTaskTemplates))
}

//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func (h *Handler) registerTaskRoutes(m *http.ServeMux) {
	m.Handle("POST /tasks", http.HandlerFunc(h.create))
	m.Handle("GET /tasks", http.HandlerFunc(h.list))
	m.Handle("GET /tasks/stats", http.HandlerFunc(h.stats))
	m.Handle("GET /tasks/cycle-failures", http.HandlerFunc(h.cycleFailures))
	m.Handle("GET /tasks/{id}/checklist", http.HandlerFunc(h.getChecklist))
	m.Handle("POST /tasks/{id}/checklist/items", http.HandlerFunc(h.postChecklistItem))
	m.Handle("PATCH /tasks/{id}/checklist/items/{itemId}", http.HandlerFunc(h.patchChecklistItem))
	m.Handle("DELETE /tasks/{id}/checklist/items/{itemId}", http.HandlerFunc(h.deleteChecklistItem))
	m.Handle("GET /tasks/{id}/events/{seq}", http.HandlerFunc(h.taskEvent))
	m.Handle("PATCH /tasks/{id}/events/{seq}", http.HandlerFunc(h.patchTaskEventUserResponse))
	m.Handle("GET /tasks/{id}/events", http.HandlerFunc(h.taskEvents))
	m.Handle("POST /tasks/{id}/cycles", http.HandlerFunc(h.postTaskCycle))
	m.Handle("GET /tasks/{id}/cycles", http.HandlerFunc(h.getTaskCycles))
	m.Handle("GET /tasks/{id}/cycles/{cycleId}/stream", http.HandlerFunc(h.getTaskCycleStream))
	m.Handle("GET /tasks/{id}/commits", http.HandlerFunc(h.getTaskCommits))
	m.Handle("GET /tasks/{id}/cycles/{cycleId}/verdicts", http.HandlerFunc(h.getTaskCycleVerdicts))
	m.Handle("GET /tasks/{id}/cycles/{cycleId}", http.HandlerFunc(h.getTaskCycle))
	m.Handle("PATCH /tasks/{id}/cycles/{cycleId}", http.HandlerFunc(h.patchTaskCycle))
	m.Handle("POST /tasks/{id}/cycles/{cycleId}/phases", http.HandlerFunc(h.postTaskCyclePhase))
	m.Handle("PATCH /tasks/{id}/cycles/{cycleId}/phases/{phaseSeq}", http.HandlerFunc(h.patchTaskCyclePhase))
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
func (h *Handler) registerRepoRoutes(m *http.ServeMux) {
	m.Handle("GET /repo/search", http.HandlerFunc(h.repoSearch))
	m.Handle("GET /repo/file", http.HandlerFunc(h.repoFile))
	m.Handle("GET /repo/validate-range", http.HandlerFunc(h.repoValidateRange))
	m.Handle("GET /repo/diff", http.HandlerFunc(h.repoDiff))
}

//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func (h *Handler) registerRunnerRoutes(m *http.ServeMux) {
	m.Handle("GET /runners", http.HandlerFunc(h.listRunners))
	m.Handle("GET /runners/{id}/config-schema", http.HandlerFunc(h.runnerConfigSchema))
	m.Handle("POST /runners/{id}/probe", http.HandlerFunc(h.probeRunner))
	m.Handle("POST /runners/{id}/list-models", http.HandlerFunc(h.listRunnerModels))
	m.Handle("POST /runners/{id}/validate-config", http.HandlerFunc(h.validateRunnerConfig))
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
