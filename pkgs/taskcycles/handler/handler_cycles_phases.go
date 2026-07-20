package handler

import (
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

// postTaskCyclePhase handles POST /tasks/{id}/cycles/{cycleId}/phases.
func (h *Handler) postTaskCyclePhase(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.postTaskCyclePhase")
	const op = "tasks.cycle.phase.create"
	r = calltrace.WithRequestRoot(r, op)
	taskID, cycleID, err := parseCyclePathPair(r)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	var body phaseStartJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.DebugHTTPRequest(r, op, "task_id", taskID, "cycle_id", cycleID, "json_decode_failed", true)
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	by := handlerhttp.ActorFromRequest(r)
	handlerhttp.DebugHTTPRequest(r, op, "task_id", taskID, "cycle_id", cycleID,
		"actor", string(by), "body_phase", string(body.Phase))
	if err := assertCycleBelongsToTask(r.Context(), h.cycles, taskID, cycleID); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	phase, err := h.cycles.StartPhase(r.Context(), cycleID, body.Phase, by)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyCycleChangedFromStore(r.Context(), taskID, cycleID)
	handlerhttp.WriteJSON(w, r, op, http.StatusCreated, taskCyclePhaseResponseFromDomain(phase))
}

// patchTaskCyclePhase handles PATCH /tasks/{id}/cycles/{cycleId}/phases/{phaseSeq}.
func (h *Handler) patchTaskCyclePhase(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.patchTaskCyclePhase")
	const op = "tasks.cycle.phase.complete"
	r = calltrace.WithRequestRoot(r, op)
	taskID, cycleID, err := parseCyclePathPair(r)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	phaseSeq, err := parseTaskPathPhaseSeq(r.PathValue("phaseSeq"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	var body phasePatchJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.DebugHTTPRequest(r, op, "task_id", taskID, "cycle_id", cycleID, "phase_seq", phaseSeq, "json_decode_failed", true)
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	by := handlerhttp.ActorFromRequest(r)
	handlerhttp.DebugHTTPRequest(r, op, "task_id", taskID, "cycle_id", cycleID, "phase_seq", phaseSeq,
		"actor", string(by), "body_status", string(body.Status),
		"summary_set", body.Summary != nil, "details_bytes", len(body.Details))
	if err := assertCycleBelongsToTask(r.Context(), h.cycles, taskID, cycleID); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	in := contract.CompletePhaseInput{
		CycleID:  cycleID,
		PhaseSeq: phaseSeq,
		Status:   body.Status,
		Summary:  body.Summary,
		Details:  []byte(body.Details),
		By:       by,
	}
	ph, err := h.cycles.CompletePhase(r.Context(), in)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyCycleChangedFromStore(r.Context(), taskID, cycleID)
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, taskCyclePhaseResponseFromDomain(ph))
}
