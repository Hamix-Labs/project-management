package handlerhttp

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
)

// MaxHTTPLogQueryBytes caps RawQuery in DebugHTTPRequest.
const MaxHTTPLogQueryBytes = 1024

// MaxHTTPLogTextRunes caps free-text previews in http.io debug fields
// (checklist text, event user_response, cycle terminate reason, etc.).
const MaxHTTPLogTextRunes = 240

// MaxHTTPLogTitleRunes caps short path/query/title previews in http.io logs.
const MaxHTTPLogTitleRunes = 160

// DebugHTTPRequest logs structured request context (method, path, query, headers safe for logs).
// Skips work when Debug is disabled for ctx.
func DebugHTTPRequest(r *http.Request, op string, extra ...any) {
	if r == nil || !slog.Default().Enabled(r.Context(), slog.LevelDebug) {
		return
	}
	q := r.URL.RawQuery
	if len(q) > MaxHTTPLogQueryBytes {
		q = apijson.TruncateUTF8ByBytes(q, MaxHTTPLogQueryBytes)
	}
	args := []any{
		"cmd", calltrace.LogCmd,
		"obs_category", "http_io",
		"operation", op,
		"call_path", calltrace.Path(r.Context()),
		"phase", "in",
		"method", r.Method,
		"path", r.URL.Path,
		"route_pattern", r.Pattern,
		"query", q,
		"content_length", r.ContentLength,
		"x_actor", strings.TrimSpace(r.Header.Get("X-Actor")),
	}
	args = append(args, extra...)
	slog.Log(r.Context(), slog.LevelDebug, "http.io", args...)
}

// DebugHTTPOut logs a non-JSON outcome (e.g. 204) at Debug.
func DebugHTTPOut(ctx context.Context, op string, httpStatus int, extra ...any) {
	if ctx == nil || !slog.Default().Enabled(ctx, slog.LevelDebug) {
		return
	}
	args := []any{
		"cmd", calltrace.LogCmd,
		"obs_category", "http_io",
		"operation", op,
		"call_path", calltrace.Path(ctx),
		"phase", "out",
		"http_status", httpStatus,
	}
	args = append(args, extra...)
	slog.Log(ctx, slog.LevelDebug, "http.io", args...)
}

// TruncateRunes limits s to maxRunes runes and appends "…" when truncated.
// Pure helper used from debug http.io field builders only.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func TruncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= maxRunes {
			b.WriteString("…")
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}
