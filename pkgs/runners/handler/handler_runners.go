package handler

import (
	"encoding/json"
	"errors"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/registry"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

// runnerProbeTimeout caps POST /runners/{id}/probe.
const runnerProbeTimeout = 5 * time.Second

// runnerListModelsTimeout caps POST /runners/{id}/list-models.
const runnerListModelsTimeout = 30 * time.Second

// ---------------------------------------------------------------------------
// GET /runners
// ---------------------------------------------------------------------------

type runnerDescriptorWire struct {
	ID                string               `json:"id"`
	Label             string               `json:"label"`
	DefaultBinaryHint string               `json:"default_binary_hint"`
	ConfigSchema      *runner.ConfigSchema `json:"config_schema,omitempty"`
}

func (h *Handler) listRunners(w http.ResponseWriter, r *http.Request) {
	const op = "runners.list"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runners.handler.Handler.listRunners")
	r = calltrace.WithRequestRoot(r, op)
	debugHTTPRequest(r, op)

	descs := registry.List()
	out := make([]runnerDescriptorWire, 0, len(descs))
	for _, d := range descs {
		wire := runnerDescriptorWire{
			ID:                d.ID,
			Label:             d.Label,
			DefaultBinaryHint: d.DefaultBinaryHint,
		}
		built, err := registry.Build(d.ID, registry.BuildOptions{})
		if err == nil {
			if csp, ok := built.(runner.ConfigSchemaProvider); ok {
				schema := csp.ConfigSchema()
				wire.ConfigSchema = &schema
			}
		}
		out = append(out, wire)
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, out)
}

// ---------------------------------------------------------------------------
// POST /runners/{id}/probe
// ---------------------------------------------------------------------------

type runnerProbeResponse struct {
	OK         bool   `json:"ok"`
	Runner     string `json:"runner"`
	BinaryPath string `json:"binary_path,omitempty"`
	Version    string `json:"version,omitempty"`
	Error      string `json:"error,omitempty"`
}

func (h *Handler) probeRunner(w http.ResponseWriter, r *http.Request) {
	const op = "runners.probe"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runners.handler.Handler.probeRunner")
	r = calltrace.WithRequestRoot(r, op)
	debugHTTPRequest(r, op)

	runnerID := r.PathValue("id")
	if runnerID == "" {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "runner id required")
		return
	}

	var body struct {
		BinaryPath string `json:"binary_path,omitempty"`
	}
	if r.ContentLength != 0 {
		if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
			if !errors.Is(err, io.EOF) {
				handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
				return
			}
		}
	}
	body.BinaryPath = strings.TrimSpace(body.BinaryPath)

	if body.BinaryPath == "" {
		cfg, err := h.settings.GetSettings(r.Context())
		if err != nil {
			handlerhttp.WriteStoreError(w, r, op, err)
			return
		}
		if runnerID == registry.CursorRunnerID {
			body.BinaryPath = cfg.CursorBin
		}
	}

	version, resolvedBin, err := registry.Probe(r.Context(), runnerID, body.BinaryPath, runnerProbeTimeout)
	resp := runnerProbeResponse{Runner: runnerID, BinaryPath: resolvedBin}
	if err != nil {
		resp.OK = false
		resp.Error = err.Error()
		if errors.Is(err, registry.ErrUnknownRunner) {
			handlerhttp.WriteJSON(w, r, op, http.StatusNotFound, resp)
			return
		}
		if errors.Is(err, runner.ErrCapabilityNotSupported) {
			handlerhttp.WriteJSON(w, r, op, http.StatusNotImplemented, resp)
			return
		}
		handlerhttp.WriteJSON(w, r, op, http.StatusOK, resp)
		return
	}
	resp.OK = true
	resp.Version = version
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, resp)
}

// ---------------------------------------------------------------------------
// POST /runners/{id}/list-models
// ---------------------------------------------------------------------------

type runnerListModelsResponse struct {
	OK         bool               `json:"ok"`
	Runner     string             `json:"runner"`
	BinaryPath string             `json:"binary_path,omitempty"`
	Models     []runner.ModelInfo `json:"models,omitempty"`
	Error      string             `json:"error,omitempty"`
}

func (h *Handler) listRunnerModels(w http.ResponseWriter, r *http.Request) {
	const op = "runners.list_models"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runners.handler.Handler.listRunnerModels")
	r = calltrace.WithRequestRoot(r, op)
	debugHTTPRequest(r, op)

	runnerID := r.PathValue("id")
	if runnerID == "" {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "runner id required")
		return
	}

	var body struct {
		BinaryPath string `json:"binary_path,omitempty"`
	}
	if r.ContentLength != 0 {
		if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
			if !errors.Is(err, io.EOF) {
				handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
				return
			}
		}
	}
	body.BinaryPath = strings.TrimSpace(body.BinaryPath)

	if body.BinaryPath == "" {
		cfg, err := h.settings.GetSettings(r.Context())
		if err != nil {
			handlerhttp.WriteStoreError(w, r, op, err)
			return
		}
		if runnerID == registry.CursorRunnerID {
			body.BinaryPath = cfg.CursorBin
		}
	}

	models, resolvedBin, err := registry.ListModelsForRunner(r.Context(), runnerID, body.BinaryPath, runnerListModelsTimeout)
	resp := runnerListModelsResponse{Runner: runnerID, BinaryPath: resolvedBin}
	if err != nil {
		resp.OK = false
		resp.Error = err.Error()
		if errors.Is(err, registry.ErrUnknownRunner) {
			handlerhttp.WriteJSON(w, r, op, http.StatusNotFound, resp)
			return
		}
		if errors.Is(err, runner.ErrCapabilityNotSupported) {
			handlerhttp.WriteJSON(w, r, op, http.StatusNotImplemented, resp)
			return
		}
		handlerhttp.WriteJSON(w, r, op, http.StatusOK, resp)
		return
	}
	resp.OK = true
	resp.Models = models
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, resp)
}

// ---------------------------------------------------------------------------
// GET /runners/{id}/config-schema
// ---------------------------------------------------------------------------

func (h *Handler) runnerConfigSchema(w http.ResponseWriter, r *http.Request) {
	const op = "runners.config_schema"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runners.handler.Handler.runnerConfigSchema")
	r = calltrace.WithRequestRoot(r, op)
	debugHTTPRequest(r, op)

	runnerID := r.PathValue("id")
	if runnerID == "" {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "runner id required")
		return
	}

	built, err := registry.Build(runnerID, registry.BuildOptions{})
	if err != nil {
		if errors.Is(err, registry.ErrUnknownRunner) {
			handlerhttp.WriteJSONError(w, r, op, http.StatusNotFound, "unknown runner")
			return
		}
		slog.ErrorContext(r.Context(), "runner registry build failed",
			"cmd", calltrace.LogCmd, "operation", op, "runner_id", runnerID, "err", err)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "failed to load runner")
		return
	}
	csp, ok := built.(runner.ConfigSchemaProvider)
	if !ok {
		handlerhttp.WriteJSONError(w, r, op, http.StatusNotImplemented, "runner does not expose a config schema")
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, csp.ConfigSchema())
}

// ---------------------------------------------------------------------------
// POST /runners/{id}/validate-config
// ---------------------------------------------------------------------------

func (h *Handler) validateRunnerConfig(w http.ResponseWriter, r *http.Request) {
	const op = "runners.validate_config"
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "runners.handler.Handler.validateRunnerConfig")
	r = calltrace.WithRequestRoot(r, op)
	debugHTTPRequest(r, op)

	runnerID := r.PathValue("id")
	if runnerID == "" {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "runner id required")
		return
	}

	built, err := registry.Build(runnerID, registry.BuildOptions{})
	if err != nil {
		if errors.Is(err, registry.ErrUnknownRunner) {
			handlerhttp.WriteJSONError(w, r, op, http.StatusNotFound, "unknown runner")
			return
		}
		slog.ErrorContext(r.Context(), "runner registry build failed",
			"cmd", calltrace.LogCmd, "operation", op, "runner_id", runnerID, "err", err)
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "failed to load runner")
		return
	}
	cv, ok := built.(runner.ConfigValidator)
	if !ok {
		handlerhttp.WriteJSONError(w, r, op, http.StatusNotImplemented, "runner does not support config validation")
		return
	}

	var blob json.RawMessage
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &blob); err != nil {
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	if err := cv.ValidateConfig(blob); err != nil {
		handlerhttp.WriteJSON(w, r, op, http.StatusUnprocessableEntity, map[string]any{
			"valid": false,
			"error": err.Error(),
		})
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, map[string]any{"valid": true})
}
