package handler

import (
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// postTaskApprove handles POST /tasks/{id}/approve for human completion after pr_ready.
func (h *Handler) postTaskApprove(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.postTaskApprove")
	const op = "tasks.approve"
	r = calltrace.WithRequestRoot(r, op)
	taskID, err := h.parseTaskPathID(r.PathValue("id"))
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	by := h.httpPort.ActorFromRequest(r)
	if by != domain.ActorUser {
		h.httpPort.WriteStoreError(w, r, op, domain.ErrInvalidInput)
		return
	}
	h.debugHTTPRequest(r, op, "task_id", taskID)
	t, err := h.tasks.RequestTaskApprove(r.Context(), taskID, by)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyTaskChangedSafe(contract.ChangeTaskUpdated, taskID, t)
	taskapiDomainTasksUpdatedTotal.Inc()
	h.httpPort.WriteJSON(w, r, op, http.StatusOK, t)
}
