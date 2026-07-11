package handler

import (
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

func (h *Handler) listTaskDrafts(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.listTaskDrafts")
	const op = "task_drafts.list"
	r = calltrace.WithRequestRoot(r, op)
	limit, err := parseBoundedLimit(r.URL.Query(), 50, 100)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	rows, err := h.compose.ListDrafts(r.Context(), limit)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, map[string]any{"drafts": rows})
}

func (h *Handler) saveTaskDraft(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.saveTaskDraft")
	const op = "task_drafts.save"
	r = calltrace.WithRequestRoot(r, op)
	var body taskDraftSaveJSON
	if err := decodeJSON(r.Context(), r.Body, &body); err != nil {
		writeError(w, r, op, err, http.StatusBadRequest)
		return
	}
	saved, err := h.compose.SaveDraft(r.Context(), body.ID, body.Name, body.Payload)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusCreated, saved)
}

func (h *Handler) getTaskDraft(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.getTaskDraft")
	const op = "task_drafts.get"
	r = calltrace.WithRequestRoot(r, op)
	getNamedPayload(w, r, op, h.compose.GetDraft)
}

func (h *Handler) deleteTaskDraft(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.deleteTaskDraft")
	const op = "task_drafts.delete"
	r = calltrace.WithRequestRoot(r, op)
	deleteNamedPayload(w, r, op, h.compose.DeleteDraft)
}
