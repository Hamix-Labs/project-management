package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func (h *Handler) listTaskDependencies(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.listTaskDependencies")
	const op = "tasks.dependencies.list"
	r = calltrace.WithRequestRoot(r, op)
	id, err := h.parseTaskPathID(r.PathValue("id"))
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	if _, err := h.tasks.Get(r.Context(), id); err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	deps, err := h.tasks.ListTaskDependencies(r.Context(), id)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	if deps == nil {
		deps = []domain.DependencyEdge{}
	}
	h.httpPort.WriteJSONWithETag(w, r, op, http.StatusOK, taskDependenciesListResponse{DependsOn: deps})
}

func (h *Handler) addTaskDependency(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.addTaskDependency")
	const op = "tasks.dependencies.create"
	r = calltrace.WithRequestRoot(r, op)
	id, err := h.parseTaskPathID(r.PathValue("id"))
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	var body taskDependencyCreateJSON
	if err := h.httpPort.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		h.httpPort.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	depID := strings.TrimSpace(body.DependsOnTaskID)
	if depID == "" {
		h.httpPort.WriteStoreError(w, r, op, fmt.Errorf("%w: depends_on_task_id required", domain.ErrInvalidInput))
		return
	}
	if !domain.ValidDependencySatisfies(body.Satisfies) && body.Satisfies != "" {
		h.httpPort.WriteStoreError(w, r, op, fmt.Errorf("%w: invalid satisfies", domain.ErrInvalidInput))
		return
	}
	satisfies := domain.NormalizeDependencySatisfies(body.Satisfies)
	if err := h.tasks.AddTaskDependency(r.Context(), id, depID, satisfies); err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyChangeSafe(contract.ChangeTaskDependencyChanged, id)
	deps, err := h.tasks.ListTaskDependencies(r.Context(), id)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	if deps == nil {
		deps = []domain.DependencyEdge{}
	}
	h.httpPort.WriteJSON(w, r, op, http.StatusCreated, taskDependenciesListResponse{DependsOn: deps})
}

func (h *Handler) removeTaskDependency(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.removeTaskDependency")
	const op = "tasks.dependencies.delete"
	r = calltrace.WithRequestRoot(r, op)
	id, err := h.parseTaskPathID(r.PathValue("id"))
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	depID, err := h.parseTaskPathID(r.PathValue("depId"))
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	if err := h.tasks.RemoveTaskDependency(r.Context(), id, depID); err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyChangeSafe(contract.ChangeTaskDependencyChanged, id)
	w.WriteHeader(http.StatusNoContent)
}
