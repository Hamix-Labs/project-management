package handler

import (
	"log/slog"
	"net/http"

	settingshandler "github.com/AlexsanderHamir/Hamix/pkgs/settings/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/readpolicy"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/service"
)

// bootstrapTasksPayload mirrors listResponse so the SPA can seed
// taskQueryKeys.list directly from bootstrap without a follow-up
// GET /tasks call. The wire shape is identical on purpose.
type bootstrapTasksPayload = listResponse

// bootstrapDraftsPayload mirrors the GET /task-drafts envelope so the
// SPA's existing draft list parser consumes it unchanged.
type bootstrapDraftsPayload struct {
	Drafts any `json:"drafts"`
}

// bootstrapResponse is the aggregate payload for GET /v1/bootstrap.
// Each field corresponds to one of the cold-start fetches the SPA used
// to fan out at App mount (settings, root list, stats, projects,
// drafts). Sub-call failure aborts the whole response with 5xx —
// partial bootstrap is more painful for the client to handle than
// falling back to the per-endpoint fan-out it already has to support.
//
// This endpoint is intentionally an *optimization hint*, not the
// canonical shape for the listed resources. Clients that want the
// precise per-endpoint guarantees should keep using GET /tasks,
// GET /settings, etc. — bootstrap clients must tolerate its absence
// (older or stripped-down servers) and gracefully fall back.
type bootstrapResponse struct {
	Settings settingshandler.SettingsWireResponse `json:"settings"`
	Tasks    bootstrapTasksPayload                `json:"tasks"`
	Stats    taskStatsResponse                    `json:"stats"`
	Projects projectsListResponse                 `json:"projects"`
	Drafts   bootstrapDraftsPayload               `json:"drafts"`
}

// bootstrap serves GET /v1/bootstrap. It composes the five cold-start
// reads in parallel via errgroup so the SPA can seed its TanStack
// Query cache from a single round trip. Any sub-call failure aborts
// the whole response with 5xx and the client falls back to its
// per-endpoint fan-out.
func (h *Handler) bootstrap(w http.ResponseWriter, r *http.Request) {
	const op = "bootstrap.aggregate"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.bootstrap")
	r = calltrace.WithRequestRoot(r, op)
	debugHTTPRequest(r, op)

	ctx := r.Context()
	data, err := service.Bootstrap(ctx, h.store, service.BootstrapLimits{
		TasksLimit:    readpolicy.BootstrapListLimit,
		ProjectsLimit: readpolicy.BootstrapProjectsLimit,
		DraftsLimit:   readpolicy.BootstrapDraftsLimit,
	})
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}

	resp := bootstrapResponse{
		Settings: settingshandler.SettingsWireFrom(data.Settings),
		Tasks:    buildListResponse(data.Tasks, readpolicy.BootstrapListLimit, 0, data.HasMore),
		Stats:    taskStatsResponseFromStore(data.Stats),
		Projects: projectsListResponse{
			Projects: data.Projects,
			Limit:    readpolicy.BootstrapProjectsLimit,
		},
		Drafts: bootstrapDraftsPayload{Drafts: data.Drafts},
	}
	writeJSONWithETag(w, r, op, http.StatusOK, resp)
}
