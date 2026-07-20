package handler

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"log/slog"
	"net/http"

	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
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
	taskID, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	by := handlerhttp.ActorFromRequest(r)
	if by != domain.ActorUser {
		handlerhttp.WriteStoreError(w, r, op, domain.ErrInvalidInput)
		return
	}
	var body taskRetryJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		debugHTTPRequest(r, op, "task_id", taskID, "json_decode_failed", true)
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	debugHTTPRequest(r, op, "task_id", taskID, "mode", string(body.Mode), "parent_cycle_id", body.ParentCycleID)
	t, err := h.tasks.RequestTaskRetry(r.Context(), taskcorecontract.RequestRetryInput{
		TaskID:        taskID,
		Mode:          body.Mode,
		ParentCycleID: body.ParentCycleID,
	}, by)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyTaskChangedSafe(realtime.TaskUpdated, taskID, t)
	taskapiDomainTasksUpdatedTotal.Inc()
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, t)
}
