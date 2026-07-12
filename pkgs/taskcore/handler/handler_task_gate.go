package handler

import (
	"fmt"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func (h *Handler) patchTaskGate(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.patchTaskGate")
	const op = "tasks.gate.patch"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	var body taskGateActionJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	action := strings.TrimSpace(strings.ToLower(body.Action))
	if action == "" {
		handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: action required", domain.ErrInvalidInput))
		return
	}
	by := handlerhttp.ActorFromRequest(r)
	if _, err := h.tasks.ApplyTaskGateAction(r.Context(), id, action, by); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	task, err := h.tasks.Get(r.Context(), id)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyChangeSafe(realtime.TaskGateChanged, id)
	h.notifyTaskChangedSafe(realtime.TaskUpdated, id, task)
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, task)
}
