package handler

import (
	"encoding/json"
	"fmt"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func (h *Handler) listTaskTemplates(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.handler.listTaskTemplates")
	const op = "task_templates.list"
	r = calltrace.WithRequestRoot(r, op)
	limit, err := handlerhttp.ParseBoundedLimit(r.URL.Query(), 50, 100)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	sort, order, tag, err := parseTemplateListQuery(r.URL.Query())
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	rows, err := h.compose.ListTemplates(r.Context(), contract.ListTemplatesInput{
		Limit: limit, Q: q, Sort: sort, Order: order, Tag: tag,
	})
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, map[string]any{"templates": rows})
}

func (h *Handler) saveTaskTemplate(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.handler.saveTaskTemplate")
	const op = "task_templates.save"
	r = calltrace.WithRequestRoot(r, op)
	var body taskTemplateSaveJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	if h.normalizeCompose == nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "internal server error")
		return
	}
	normalized, err := h.normalizeCompose(r.Context(), body.Payload)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = strings.TrimSpace(normalized.Title)
	}
	saved, err := h.compose.SaveTemplate(r.Context(), body.ID, name, normalized.Payload)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusCreated, saved)
}

func (h *Handler) getTaskTemplate(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.handler.getTaskTemplate")
	const op = "task_templates.get"
	r = calltrace.WithRequestRoot(r, op)
	getNamedPayload(w, r, op, h.compose.GetTemplate)
}

func (h *Handler) patchTaskTemplate(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.handler.patchTaskTemplate")
	const op = "task_templates.patch"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	var body taskTemplatePatchJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	var payloadRaw json.RawMessage
	if len(body.Payload) > 0 {
		if h.normalizeCompose == nil {
			handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "internal server error")
			return
		}
		normalized, nerr := h.normalizeCompose(r.Context(), body.Payload)
		if nerr != nil {
			handlerhttp.WriteStoreError(w, r, op, nerr)
			return
		}
		payloadRaw = normalized.Payload
	}
	name := body.Name
	if name == nil && payloadRaw == nil {
		handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: no fields to update", taskcoredomain.ErrInvalidInput))
		return
	}
	updated, err := h.compose.PatchTemplate(r.Context(), id, name, payloadRaw)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, updated)
}

func (h *Handler) deleteTaskTemplate(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.handler.deleteTaskTemplate")
	const op = "task_templates.delete"
	r = calltrace.WithRequestRoot(r, op)
	deleteNamedPayload(w, r, op, h.compose.DeleteTemplate)
}

func (h *Handler) instantiateTaskTemplates(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcompose.handler.instantiateTaskTemplates")
	const op = "task_templates.instantiate"
	r = calltrace.WithRequestRoot(r, op)
	var body taskTemplateInstantiateJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	items, err := normalizeInstantiateItems(body)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	if h.instantiateFromTemplate == nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "internal server error")
		return
	}
	by := handlerhttp.ActorFromRequest(r)
	resp := taskTemplateInstantiateResponseJSON{
		Tasks:  make([]taskcoredomain.Task, 0),
		Errors: make([]taskTemplateInstantiateErrorJSON, 0),
	}
	successCounts := make(map[string]int)
	for _, item := range items {
		detail, err := h.compose.GetTemplate(r.Context(), item.TemplateID)
		if err != nil {
			resp.Errors = append(resp.Errors, taskTemplateInstantiateErrorJSON{
				TemplateID: item.TemplateID,
				Error:      err.Error(),
			})
			continue
		}
		payloadRaw := detail.Payload
		applied, applyErr := applyFunctionBindingsToPayload(detail.Payload, item.FunctionBindings)
		if applyErr != nil {
			resp.Errors = append(resp.Errors, taskTemplateInstantiateErrorJSON{
				TemplateID: item.TemplateID,
				Error:      applyErr.Error(),
			})
			continue
		}
		payloadRaw = applied
		for range item.Count {
			task, err := h.instantiateFromTemplate(r.Context(), r, op, payloadRaw, by)
			if err != nil {
				resp.Errors = append(resp.Errors, taskTemplateInstantiateErrorJSON{
					TemplateID: item.TemplateID,
					Error:      err.Error(),
				})
				continue
			}
			resp.Tasks = append(resp.Tasks, *task)
			successCounts[item.TemplateID]++
		}
	}
	if len(successCounts) > 0 {
		if err := h.compose.IncrementTemplateInstantiateCounts(r.Context(), successCounts); err != nil {
			handlerhttp.WriteStoreError(w, r, op, err)
			return
		}
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, resp)
}
