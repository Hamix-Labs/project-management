package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/httperr"
)

// Shared JSON/error paths live in handlerhttp / httperr / apijson (B-21).
// These thin wrappers keep gitinventory handler call sites concise.

func decodeJSON(ctx context.Context, r io.Reader, dst any) error {
	return handlerhttp.DecodeJSON(ctx, r, dst)
}

func writeJSON(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	handlerhttp.WriteJSON(w, r, op, code, v)
}

func writeError(w http.ResponseWriter, r *http.Request, op string, err error, code int) {
	handlerhttp.WriteError(w, r, op, err, code)
}

// WriteGitStoreError writes a git-aware JSON error response.
func WriteGitStoreError(w http.ResponseWriter, r *http.Request, op string, err error) {
	httperr.WriteGitStoreError(w, r, op, err)
}

// GitErrHTTP maps git and domain store errors to HTTP status, code, and message.
func GitErrHTTP(err error) (status int, code, msg string) {
	return httperr.GitErrHTTP(err)
}

func gitWorktreeDeleteQuery(r *http.Request, op string) (removeFromDisk, force bool) {
	removeFromDisk = r.URL.Query().Get("remove_from_disk") == "true"
	force = r.URL.Query().Get("force") == "true"
	if force && !removeFromDisk {
		slog.Warn("deprecated query param ignored", "cmd", calltrace.LogCmd, "operation", op, "param", "force")
	}
	return removeFromDisk, force
}
