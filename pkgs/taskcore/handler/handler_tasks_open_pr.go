package handler

import (
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

type taskOpenPRJSON struct {
	ParentCycleID string `json:"parent_cycle_id,omitempty"`
}

// postTaskOpenPR handles POST /tasks/{id}/open-pr for approve-and-open-PR from review.
func (h *Handler) postTaskOpenPR(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.postTaskOpenPR")
	const op = "tasks.open_pr"
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
	var body taskOpenPRJSON
	if err := h.httpPort.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		h.debugHTTPRequest(r, op, "task_id", taskID, "json_decode_failed", true)
		h.httpPort.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	h.debugHTTPRequest(r, op, "task_id", taskID, "parent_cycle_id", body.ParentCycleID)
	t, err := h.tasks.RequestTaskOpenPR(r.Context(), contract.RequestOpenPRInput{
		TaskID:        taskID,
		ParentCycleID: body.ParentCycleID,
	}, by)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyTaskChangedSafe(contract.ChangeTaskUpdated, taskID, t)
	taskapiDomainTasksUpdatedTotal.Inc()
	h.httpPort.WriteJSON(w, r, op, http.StatusOK, t)
}
