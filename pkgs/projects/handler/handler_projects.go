package handler

import (
	"fmt"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/projects/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.createProject")
	const op = "projects.create"
	r = calltrace.WithRequestRoot(r, op)
	var body projectCreateJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	project, err := h.store.CreateProject(r.Context(), contract.CreateProjectInput{
		ID:             body.ID,
		Name:           body.Name,
		Description:    body.Description,
		ContextSummary: body.ContextSummary,
		RepositoryID:   body.RepositoryID,
	})
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	h.notifyChange(realtime.ProjectCreated, project.ID)
	handlerhttp.WriteJSON(w, r, op, http.StatusCreated, project)
}

func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.listProjects")
	const op = "projects.list"
	r = calltrace.WithRequestRoot(r, op)
	limit, includeArchived, err := parseProjectListParams(r.URL.Query())
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	projects, err := h.store.ListProjects(r.Context(), includeArchived, limit)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSONWithETag(w, r, op, http.StatusOK, projectsListResponse{Projects: projects, Limit: limit})
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.getProject")
	const op = "projects.get"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	project, err := h.store.GetProject(r.Context(), id)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSONWithETag(w, r, op, http.StatusOK, project)
}

func (h *Handler) patchProject(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.patchProject")
	const op = "projects.patch"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	var body projectPatchJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	if body.isEmpty() {
		writeStoreError(w, r, op, fmt.Errorf("%w: no fields to update", domain.ErrInvalidInput))
		return
	}
	project, err := h.store.UpdateProject(r.Context(), id, contract.UpdateProjectInput{
		Name:           body.Name,
		Description:    body.Description,
		Status:         body.Status,
		ContextSummary: body.ContextSummary,
	})
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	h.notifyChange(realtime.ProjectUpdated, project.ID)
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, project)
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.deleteProject")
	const op = "projects.delete"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	if err := h.store.DeleteProject(r.Context(), id); err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	h.notifyChange(realtime.ProjectDeleted, id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createProjectContext(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.createProjectContext")
	const op = "projects.context.create"
	r = calltrace.WithRequestRoot(r, op)
	projectID, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	var body projectContextCreateJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	item, err := h.store.CreateProjectContext(r.Context(), projectID, contract.CreateProjectContextInput{
		ID:            body.ID,
		Kind:          body.Kind,
		Title:         body.Title,
		Body:          body.Body,
		SourceTaskID:  body.SourceTaskID,
		SourceCycleID: body.SourceCycleID,
		CreatedBy:     domain.Actor(handlerhttp.ActorFromRequest(r)),
		Pinned:        body.Pinned,
	})
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	h.notifyChange(realtime.ProjectContextChanged, projectID)
	handlerhttp.WriteJSON(w, r, op, http.StatusCreated, item)
}

func (h *Handler) listProjectContext(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.listProjectContext")
	const op = "projects.context.list"
	r = calltrace.WithRequestRoot(r, op)
	projectID, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	limit, includeUnpinned, err := parseProjectContextListParams(r.URL.Query())
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	items, err := h.store.ListProjectContext(r.Context(), projectID, includeUnpinned, limit)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	// Empty store results are nil slices; JSON would encode them as null and
	// break the SPA parser which requires items/edges arrays.
	if items == nil {
		items = []domain.ProjectContextItem{}
	}
	// Edges are no longer exposed; keep an empty array for JSON compat.
	handlerhttp.WriteJSONWithETag(w, r, op, http.StatusOK, projectContextListResponse{
		Items: items,
		Edges: []domain.ProjectContextEdge{},
		Limit: limit,
	})
}

func (h *Handler) patchProjectContext(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.patchProjectContext")
	const op = "projects.context.patch"
	r = calltrace.WithRequestRoot(r, op)
	projectID, itemID, err := parseProjectContextPath(r)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	var body projectContextPatchJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	if body.isEmpty() {
		writeStoreError(w, r, op, fmt.Errorf("%w: no fields to update", domain.ErrInvalidInput))
		return
	}
	item, err := h.store.UpdateProjectContext(r.Context(), projectID, itemID, contract.UpdateProjectContextInput{
		Kind:   body.Kind,
		Title:  body.Title,
		Body:   body.Body,
		Pinned: body.Pinned,
	})
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	h.notifyChange(realtime.ProjectContextChanged, projectID)
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, item)
}

func (h *Handler) deleteProjectContext(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.deleteProjectContext")
	const op = "projects.context.delete"
	r = calltrace.WithRequestRoot(r, op)
	projectID, itemID, err := parseProjectContextPath(r)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	if err := h.store.DeleteProjectContext(r.Context(), projectID, itemID); err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	h.notifyChange(realtime.ProjectContextChanged, projectID)
	w.WriteHeader(http.StatusNoContent)
}
