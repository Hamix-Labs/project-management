package handler

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

// getTaskTokenUsage handles GET /tasks/{id}/token-usage.
func (h *Handler) getTaskTokenUsage(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.getTaskTokenUsage")
	const op = "tasks.token_usage.get"
	r = calltrace.WithRequestRoot(r, op)
	taskID, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.DebugHTTPRequest(r, op, "task_id", taskID)
	if _, err := h.tasks.Get(r.Context(), taskID); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	rows, err := h.cycles.ListPhaseTokenUsageForTask(r.Context(), taskID)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	exec, verify := cyclesdomain.SumPhaseUsageByKind(rows)
	taskUsage := projectTokenUsage(exec, verify, len(rows) > 0)
	byCycle := groupUsageRowsByCycleID(rows)
	attempts := make([]taskTokenUsageAttempt, 0, len(byCycle))
	for cycleID, cycleRows := range byCycle {
		proj := projectTokenUsageFromRows(cycleRows)
		attempts = append(attempts, taskTokenUsageAttempt{
			CycleID:        cycleID,
			AttemptSeq:     cycleRows[0].AttemptSeq,
			TokenUsage:     proj,
			ShareOfTaskPct: shareOfTaskPct(proj.ConsumedTokens, taskUsage.ConsumedTokens, proj.Known && taskUsage.Known),
		})
	}
	sort.Slice(attempts, func(i, j int) bool {
		return attempts[i].AttemptSeq < attempts[j].AttemptSeq
	})
	resp := taskTokenUsageResponse{
		TaskID:     taskID,
		TokenUsage: taskUsage,
		Attempts:   attempts,
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, resp)
}
