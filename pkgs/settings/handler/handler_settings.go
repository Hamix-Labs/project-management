package handler

import (
	"errors"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	const op = "settings.get"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.handler.getSettings")
	r = calltrace.WithRequestRoot(r, op)
	handlerhttp.DebugHTTPRequest(r, op)

	cfg, err := h.settings.GetSettings(r.Context())
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSONWithETag(w, r, op, http.StatusOK, h.settingsResponseFrom(cfg))
}

func (h *Handler) patchSettings(w http.ResponseWriter, r *http.Request) {
	const op = "settings.patch"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.handler.patchSettings")
	r = calltrace.WithRequestRoot(r, op)
	handlerhttp.DebugHTTPRequest(r, op)

	if h.agent == nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusServiceUnavailable, "agent worker control unavailable")
		return
	}

	var body settingsPatchBody
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	patch := contract.SettingsPatch{
		AgentPaused:                body.AgentPaused,
		Runner:                     body.Runner,
		CursorBin:                  body.CursorBin,
		CursorModel:                body.CursorModel,
		VerifyModel:                body.VerifyModel,
		MaxRunDurationSeconds:      body.MaxRunDurationSeconds,
		AgentPickupDelaySeconds:    body.AgentPickupDelaySeconds,
		DisplayTimezone:            body.DisplayTimezone,
		CursorSessionResumeEnabled: body.CursorSessionResumeEnabled,
		AgentMCPEnabled:            body.AgentMCPEnabled,
	}
	if patch.IsEmpty() {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "patch body must include at least one field")
		return
	}

	updated, err := h.settings.UpdateSettings(r.Context(), patch)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	if reloadErr := h.agent.Reload(r.Context()); reloadErr != nil {
		slog.Error("settings patch persisted but supervisor reload failed",
			"cmd", calltrace.LogCmd, "operation", op, "err", reloadErr)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "settings saved but worker reload failed")
		return
	}
	h.notifyChange(realtime.SettingsChanged)
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, h.settingsResponseFrom(updated))
}

// Deprecated: use POST /runners/{id}/probe instead.
func (h *Handler) probeCursor(w http.ResponseWriter, r *http.Request) {
	const op = "settings.probe_cursor"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.handler.probeCursor")
	r = calltrace.WithRequestRoot(r, op)
	handlerhttp.DebugHTTPRequest(r, op)

	if h.agent == nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusServiceUnavailable, "agent worker control unavailable")
		return
	}

	var body probeRequest
	if r.ContentLength != 0 {
		if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
			if !errors.Is(err, io.EOF) {
				handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
				return
			}
		}
	}
	body.Runner = strings.TrimSpace(body.Runner)
	body.BinaryPath = strings.TrimSpace(body.BinaryPath)

	if body.Runner == "" || body.BinaryPath == "" {
		cfg, err := h.settings.GetSettings(r.Context())
		if err != nil {
			handlerhttp.WriteStoreError(w, r, op, err)
			return
		}
		if body.Runner == "" {
			body.Runner = cfg.Runner
		}
		if body.BinaryPath == "" {
			body.BinaryPath = cfg.CursorBin
		}
	}

	version, resolvedBin, err := h.agent.ProbeRunner(r.Context(), body.Runner, body.BinaryPath, settingsProbeTimeout)
	resp := probeResponse{Runner: body.Runner, BinaryPath: resolvedBin}
	if err != nil {
		resp.OK = false
		resp.Error = err.Error()
		handlerhttp.WriteJSON(w, r, op, http.StatusOK, resp)
		return
	}
	resp.OK = true
	resp.Version = version
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, resp)
}

// Deprecated: use POST /runners/{id}/list-models instead.
func (h *Handler) listCursorModels(w http.ResponseWriter, r *http.Request) {
	const op = "settings.list_cursor_models"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.handler.listCursorModels")
	r = calltrace.WithRequestRoot(r, op)
	handlerhttp.DebugHTTPRequest(r, op)

	var body listCursorModelsRequest
	if r.ContentLength != 0 {
		if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
			if !errors.Is(err, io.EOF) {
				handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
				return
			}
		}
	}
	body.Runner = strings.TrimSpace(body.Runner)
	body.BinaryPath = strings.TrimSpace(body.BinaryPath)

	if body.Runner == "" || body.BinaryPath == "" {
		cfg, err := h.settings.GetSettings(r.Context())
		if err != nil {
			handlerhttp.WriteStoreError(w, r, op, err)
			return
		}
		if body.Runner == "" {
			body.Runner = cfg.Runner
		}
		if body.BinaryPath == "" {
			body.BinaryPath = cfg.CursorBin
		}
	}

	if h.runnerModels == nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusServiceUnavailable, "runner model listing unavailable")
		return
	}

	models, resolved, err := h.runnerModels.ListModels(r.Context(), body.Runner, body.BinaryPath, 30*time.Second)
	out := listCursorModelsResponse{Runner: body.Runner, BinaryPath: resolved}
	if err != nil {
		out.OK = false
		out.Error = err.Error()
		handlerhttp.WriteJSON(w, r, op, http.StatusOK, out)
		return
	}
	out.OK = true
	out.Models = make([]cursorModelWire, 0, len(models))
	for _, m := range models {
		out.Models = append(out.Models, cursorModelWire{ID: m.ID, Label: m.Label})
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, out)
}

func (h *Handler) cancelCurrentRun(w http.ResponseWriter, r *http.Request) {
	const op = "settings.cancel_current_run"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.handler.cancelCurrentRun")
	r = calltrace.WithRequestRoot(r, op)
	handlerhttp.DebugHTTPRequest(r, op)

	if h.agent == nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusServiceUnavailable, "agent worker control unavailable")
		return
	}
	cancelled := h.agent.CancelCurrentRun()
	if cancelled {
		h.notifyChange(realtime.AgentRunCancelled)
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, cancelRunResponse{Cancelled: cancelled})
}

func (h *Handler) settingsResponseFrom(cfg settingsdomain.AppSettings) settingsResponse {
	return settingsResponseFrom(cfg)
}

// SettingsWireResponse is the GET /settings JSON shape for bootstrap and GET /settings.
type SettingsWireResponse = settingsResponse

// SettingsWireFrom maps domain AppSettings to the stable /settings wire shape.
func SettingsWireFrom(cfg settingsdomain.AppSettings) SettingsWireResponse {
	return settingsResponseFrom(cfg)
}

func settingsResponseFrom(cfg settingsdomain.AppSettings) settingsResponse {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.handler.settingsResponseFrom")
	resp := settingsResponse{
		AgentPaused:                cfg.AgentPaused,
		Runner:                     cfg.Runner,
		CursorBin:                  cfg.CursorBin,
		CursorModel:                cfg.CursorModel,
		VerifyModel:                cfg.VerifyModel,
		MaxRunDurationSeconds:      cfg.MaxRunDurationSeconds,
		AgentPickupDelaySeconds:    cfg.AgentPickupDelaySeconds,
		DisplayTimezone:            cfg.DisplayTimezone,
		OptimisticMutationsEnabled: cfg.OptimisticMutationsEnabled,
		SSEReplayEnabled:           cfg.SSEReplayEnabled,
		CursorSessionResumeEnabled: cfg.CursorSessionResumeEnabled,
		AgentMCPEnabled:            cfg.AgentMCPEnabled,
	}
	if !cfg.UpdatedAt.IsZero() {
		resp.UpdatedAt = cfg.UpdatedAt.UTC().Format(time.RFC3339)
	}
	return resp
}
