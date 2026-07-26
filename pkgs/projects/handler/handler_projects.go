package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
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
		ID:           body.ID,
		Name:         body.Name,
		Description:  body.Description,
		RepositoryID: body.RepositoryID,
	})
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
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
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	projects, err := h.store.ListProjects(r.Context(), includeArchived, limit)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSONWithETag(w, r, op, http.StatusOK, projectsListResponse{Projects: projects, Limit: limit})
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.getProject")
	const op = "projects.get"
	r = calltrace.WithRequestRoot(r, op)
	id, err := handlerhttp.ParsePathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	project, err := h.store.GetProject(r.Context(), id)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSONWithETag(w, r, op, http.StatusOK, project)
}

func (h *Handler) patchProject(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.patchProject")
	const op = "projects.patch"
	r = calltrace.WithRequestRoot(r, op)
	id, err := handlerhttp.ParsePathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	var body projectPatchJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	if body.isEmpty() {
		handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: no fields to update", domain.ErrInvalidInput))
		return
	}
	project, err := h.store.UpdateProject(r.Context(), id, contract.UpdateProjectInput{
		Name:        body.Name,
		Description: body.Description,
		Status:      body.Status,
	})
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyChange(realtime.ProjectUpdated, project.ID)
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, project)
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.deleteProject")
	const op = "projects.delete"
	r = calltrace.WithRequestRoot(r, op)
	id, err := handlerhttp.ParsePathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	if err := h.store.DeleteProject(r.Context(), id); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyChange(realtime.ProjectDeleted, id)
	w.WriteHeader(http.StatusNoContent)
}
