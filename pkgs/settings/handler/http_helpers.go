package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

const maxHTTPLogQueryBytes = 1024

func writeStoreError(w http.ResponseWriter, r *http.Request, op string, err error) {
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, settingsdomain.ErrInvalidInput), errors.Is(err, taskcoredomain.ErrInvalidInput):
		code = http.StatusBadRequest
	}
	msg := "internal server error"
	if code != http.StatusInternalServerError {
		if d := invalidInputDetail(err); d != "" {
			msg = d
		} else {
			msg = err.Error()
		}
	}
	handlerhttp.WriteJSONError(w, r, op, code, msg)
	if r != nil {
		ctx := r.Context()
		slog.Log(ctx, slog.LevelWarn, "request failed",
			"cmd", calltrace.LogCmd, "operation", op,
			"request_id", logctx.RequestIDFromContext(ctx),
			"http_status", code, "err", err)
	}
}

func debugHTTPRequest(r *http.Request, op string, extra ...any) {
	if r == nil || !slog.Default().Enabled(r.Context(), slog.LevelDebug) {
		return
	}
	q := r.URL.RawQuery
	if len(q) > maxHTTPLogQueryBytes {
		q = apijson.TruncateUTF8ByBytes(q, maxHTTPLogQueryBytes)
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
	slog.Debug("http request", args...)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func truncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "…"
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func invalidInputDetail(err error) string {
	s := err.Error()
	for _, mark := range []string{"settings: invalid input: ", "tasks: invalid input: "} {
		if i := strings.Index(s, mark); i >= 0 {
			return strings.TrimSpace(s[i+len(mark):])
		}
	}
	return ""
}

func repoErrUserMessage(err error) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.handler.repoErrUserMessage")
	if d := invalidInputDetail(err); d != "" {
		return d
	}
	return err.Error()
}
