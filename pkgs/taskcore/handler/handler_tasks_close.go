package handler

import (
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
)

func (h *Handler) postTaskClose(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.postTaskClose")
	const op = "tasks.close"
	r = calltrace.WithRequestRoot(r, op)
	taskID, err := h.parseTaskPathID(r.PathValue("id"))
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	by := h.httpPort.ActorFromRequest(r)
	h.debugHTTPRequest(r, op, "task_id", taskID)
	t, err := h.tasks.Close(r.Context(), taskID, by)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyTaskChangedSafe(contract.ChangeTaskUpdated, taskID, t)
	taskapiDomainTasksUpdatedTotal.Inc()
	h.httpPort.WriteJSON(w, r, op, http.StatusOK, t)
}

func (h *Handler) postTaskReopen(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.postTaskReopen")
	const op = "tasks.reopen"
	r = calltrace.WithRequestRoot(r, op)
	taskID, err := h.parseTaskPathID(r.PathValue("id"))
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	by := h.httpPort.ActorFromRequest(r)
	h.debugHTTPRequest(r, op, "task_id", taskID)
	t, err := h.tasks.Reopen(r.Context(), taskID, by)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyTaskChangedSafe(contract.ChangeTaskUpdated, taskID, t)
	taskapiDomainTasksUpdatedTotal.Inc()
	h.httpPort.WriteJSON(w, r, op, http.StatusOK, t)
}
