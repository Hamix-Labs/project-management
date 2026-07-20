package handler

import (
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

// getTaskCycleStream handles GET /tasks/{id}/cycles/{cycleId}/stream.
func (h *Handler) getTaskCycleStream(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.getTaskCycleStream")
	const op = "tasks.cycle.stream.list"
	r = calltrace.WithRequestRoot(r, op)
	taskID, cycleID, err := parseCyclePathPair(r)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	limit, err := parseCycleStreamLimit(r.Context(), r.URL.Query())
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	afterSeq, err := parseCycleStreamAfterSeq(r.Context(), r.URL.Query())
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	if err := assertCycleBelongsToTask(r.Context(), h.cycles, taskID, cycleID); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	rows, err := h.cycles.ListCycleStreamEvents(r.Context(), cycleID, afterSeq, limit+1)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	out, hasMore, next := paginateMappedRows(rows, limit, taskCycleStreamEventResponseFromDomain, func(r taskCycleStreamEventResponse) int64 {
		return r.StreamSeq
	})
	resp := taskCycleStreamListResponse{
		TaskID:       taskID,
		CycleID:      cycleID,
		Events:       out,
		Limit:        limit,
		HasMore:      hasMore,
		NextAfterSeq: next,
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, resp)
}
