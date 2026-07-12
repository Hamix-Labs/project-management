package handler

import (
	"context"
	"errors"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"log/slog"
	"net/http"
	"time"

	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres"
)

func health(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.health")
	const op = "health"
	r = calltrace.WithRequestRoot(r, op)
	debugHTTPRequest(r, op)
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": ServerVersion(),
	})
}

func healthLive(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.health.live")
	const op = "health.live"
	r = calltrace.WithRequestRoot(r, op)
	debugHTTPRequest(r, op)
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": ServerVersion(),
	})
}

func (h *Handler) healthReady(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.health.ready")
	const op = "health.ready"
	r = calltrace.WithRequestRoot(r, op)
	debugHTTPRequest(r, op)
	ctx, cancel := context.WithTimeout(r.Context(), taskcorestore.DefaultReadyTimeout)
	defer cancel()

	checks := map[string]string{}

	if err := h.store.Ready(ctx); err != nil {
		slog.Warn("readiness check failed", "cmd", calltrace.LogCmd, "operation", op, "check", "database", "err", err,
			"deadline_exceeded", errors.Is(err, context.DeadlineExceeded),
			"timeout_sec", int(taskcorestore.DefaultReadyTimeout/time.Second))
		checks["database"] = "fail"
		handlerhttp.WriteJSON(w, r, op, http.StatusServiceUnavailable, map[string]any{
			"status":  "degraded",
			"checks":  checks,
			"version": ServerVersion(),
		})
		return
	}
	checks["database"] = "ok"

	if h.schemaDrift.FailsReadiness() {
		schemaCheck := string(h.schemaDrift.Status)
		if schemaCheck == string(postgres.SchemaDriftPending) {
			schemaCheck = "pending"
		}
		slog.Warn("readiness check failed", "cmd", calltrace.LogCmd, "operation", op, "check", "schema",
			"status", h.schemaDrift.Status,
			"code_revision", h.schemaDrift.CodeRevision,
			"db_revision", h.schemaDrift.DBRevision)
		checks["schema"] = schemaCheck
		handlerhttp.WriteJSON(w, r, op, http.StatusServiceUnavailable, map[string]any{
			"status": "degraded",
			"checks": checks,
			"schema": map[string]any{
				"message":     h.schemaDrift.OperatorMessage(),
				"remediation": h.schemaDrift.Remediation(),
			},
			"version": ServerVersion(),
		})
		return
	}
	checks["schema"] = "ok"

	if !h.gitAvailable {
		slog.Warn("readiness check failed", "cmd", calltrace.LogCmd, "operation", op, "check", "git_available")
		checks["git_available"] = "fail"
		handlerhttp.WriteJSON(w, r, op, http.StatusServiceUnavailable, map[string]any{
			"status":  "degraded",
			"checks":  checks,
			"version": ServerVersion(),
		})
		return
	}
	checks["git_available"] = "ok"

	repoCount, err := h.store.CountGitRepositories(ctx)
	if err != nil {
		slog.Warn("readiness check failed", "cmd", calltrace.LogCmd, "operation", op, "check", "registered_repositories", "err", err)
		checks["registered_repositories"] = "fail"
		handlerhttp.WriteJSON(w, r, op, http.StatusServiceUnavailable, map[string]any{
			"status":  "degraded",
			"checks":  checks,
			"version": ServerVersion(),
		})
		return
	}
	if repoCount == 0 {
		slog.Warn("readiness advisory", "cmd", calltrace.LogCmd, "operation", op, "check", "registered_repositories", "count", 0)
		checks["registered_repositories"] = "warn"
	} else {
		checks["registered_repositories"] = "ok"
	}

	handlerhttp.WriteJSON(w, r, op, http.StatusOK, map[string]any{
		"status":  "ok",
		"checks":  checks,
		"version": ServerVersion(),
	})
}
