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

func (h *Handler) patchTaskGate(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.patchTaskGate")
	const op = "tasks.gate.patch"
	r = calltrace.WithRequestRoot(r, op)
	id, err := h.parseTaskPathID(r.PathValue("id"))
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	var body taskGateActionJSON
	if err := h.httpPort.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		h.httpPort.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	action := strings.TrimSpace(strings.ToLower(body.Action))
	if action == "" {
		h.httpPort.WriteStoreError(w, r, op, fmt.Errorf("%w: action required", domain.ErrInvalidInput))
		return
	}
	by := h.httpPort.ActorFromRequest(r)
	if _, err := h.tasks.ApplyTaskGateAction(r.Context(), id, action, by); err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	task, err := h.tasks.Get(r.Context(), id)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyChangeSafe(contract.ChangeTaskGateChanged, id)
	h.notifyTaskChangedSafe(contract.ChangeTaskUpdated, id, task)
	h.httpPort.WriteJSON(w, r, op, http.StatusOK, task)
}
