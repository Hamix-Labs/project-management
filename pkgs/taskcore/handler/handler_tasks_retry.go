package handler

import (
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

type taskRetryJSON struct {
	Mode          domain.RetryMode `json:"mode"`
	ParentCycleID string           `json:"parent_cycle_id,omitempty"`
}

// postTaskRetry handles POST /tasks/{id}/retry for operator retry after failure.
func (h *Handler) postTaskRetry(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.postTaskRetry")
	const op = "tasks.retry"
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
	var body taskRetryJSON
	if err := h.httpPort.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		h.debugHTTPRequest(r, op, "task_id", taskID, "json_decode_failed", true)
		h.httpPort.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	h.debugHTTPRequest(r, op, "task_id", taskID, "mode", string(body.Mode), "parent_cycle_id", body.ParentCycleID)
	t, err := h.tasks.RequestTaskRetry(r.Context(), contract.RequestRetryInput{
		TaskID:        taskID,
		Mode:          body.Mode,
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
