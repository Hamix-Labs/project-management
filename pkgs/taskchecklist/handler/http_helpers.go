package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

const (
	maxPathIDBytes             = 128
	maxHTTPLogJSONPreviewBytes = 16384
	maxHTTPLogTextRunes        = 240
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func decodeJSON(ctx context.Context, r io.Reader, dst any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("json trailing data: %w", err)
	}
	return fmt.Errorf("%w: json trailing data", domain.ErrInvalidInput)
}

//funclogmeasure:skip category=delegate-already-logs reason="JSON response helper; HTTP handler chokepoint emits trace."
func writeJSON(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	apijson.ApplySecurityHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		writeJSONError(w, r, op, http.StatusInternalServerError, "internal server error")
		return
	}
	payload := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
	w.WriteHeader(code)
	_, _ = w.Write(payload)
	_, _ = w.Write([]byte("\n"))
}

// writeJSONWithETag encodes v as JSON, attaches a strong ETag derived from the body,
// and serves 304 Not Modified when If-None-Match matches.
func writeJSONWithETag(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	apijson.ApplyRevalidatableHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	ctx := requestCtx(r)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		if r != nil {
			rid := logctx.RequestIDFromContext(ctx)
			route := requestRouteLabel(r)
			slog.Log(ctx, slog.LevelError, "response encode failed",
				"cmd", calltrace.LogCmd, "operation", op, "request_id", rid, "route", route,
				"failure_stage", "response_encode", "err", err)
		} else {
			slog.Error("response encode failed", "cmd", calltrace.LogCmd, "operation", op,
				"failure_stage", "response_encode", "err", err)
		}
		writeJSONError(w, r, op, http.StatusInternalServerError, "internal server error")
		return
	}
	payload := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
	etag := apijson.ComputeETag(payload)
	w.Header().Set("ETag", etag)

	if r != nil && apijson.IfNoneMatchMatches(r.Header.Get("If-None-Match"), etag) {
		if slog.Default().Enabled(ctx, slog.LevelDebug) {
			slog.Log(ctx, slog.LevelDebug, "http.io",
				"cmd", calltrace.LogCmd, "obs_category", "http_io", "operation", op,
				"call_path", calltrace.Path(ctx), "phase", "out",
				"http_status", http.StatusNotModified, "etag", etag, "conditional", true)
		}
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if r != nil && slog.Default().Enabled(ctx, slog.LevelDebug) {
		preview := apijson.TruncateUTF8ByBytes(string(payload), maxHTTPLogJSONPreviewBytes)
		slog.Log(ctx, slog.LevelDebug, "http.io",
			"cmd", calltrace.LogCmd, "obs_category", "http_io", "operation", op,
			"call_path", calltrace.Path(ctx), "phase", "out",
			"http_status", code, "etag", etag,
			"response_json_bytes", len(payload), "response_body", preview)
	}
	w.WriteHeader(code)
	if _, err := w.Write(payload); err != nil {
		logResponseWriteFailure(ctx, r, op, err, "body")
		return
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		logResponseWriteFailure(ctx, r, op, err, "newline")
		return
	}
}

//funclogmeasure:skip category=delegate-already-logs reason="Error response helper; HTTP handler chokepoint emits trace."
func writeJSONError(w http.ResponseWriter, r *http.Request, op string, code int, msg string) {
	apijson.WriteJSONError(w, r, op, code, msg, calltrace.Path)
}

//funclogmeasure:skip category=delegate-already-logs reason="Error response helper; HTTP handler chokepoint emits trace."
func writeError(w http.ResponseWriter, r *http.Request, op string, err error, code int) {
	msg := http.StatusText(code)
	if code == http.StatusBadRequest {
		msg = userFacingJSONError(err)
		if msg == "" {
			msg = "bad request"
		}
	}
	writeJSONError(w, r, op, code, msg)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func userFacingJSONError(err error) string {
	s := err.Error()
	if strings.HasPrefix(s, "json decode: ") {
		return strings.TrimPrefix(s, "json decode: ")
	}
	if errors.Is(err, domain.ErrInvalidInput) {
		return "request body must contain a single JSON value"
	}
	if strings.HasPrefix(s, "json trailing data:") {
		return "request body must contain a single JSON value"
	}
	return s
}

//funclogmeasure:skip category=delegate-already-logs reason="Error response helper; HTTP handler chokepoint emits trace."
func writeStoreError(w http.ResponseWriter, r *http.Request, op string, err error) {
	code, msg := storeErrHTTP(err)
	if code >= 500 {
		slog.Log(r.Context(), slog.LevelError, "request failed",
			"cmd", calltrace.LogCmd, "operation", op,
			"request_id", logctx.RequestIDFromContext(r.Context()),
			"http_status", code, "err", err,
		)
	} else {
		slog.Log(r.Context(), slog.LevelWarn, "request failed",
			"cmd", calltrace.LogCmd, "operation", op,
			"request_id", logctx.RequestIDFromContext(r.Context()),
			"http_status", code, "err", err,
		)
	}
	writeJSONError(w, r, op, code, msg)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func storeErrHTTP(err error) (code int, msg string) {
	code = http.StatusInternalServerError
	msg = "internal server error"
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "request timed out"
	case errors.Is(err, context.Canceled):
		return http.StatusRequestTimeout, "request canceled"
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, invalidInputDetail(err)
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, conflictDetail(err)
	default:
		return code, msg
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func invalidInputDetail(err error) string {
	s := err.Error()
	const mark = "tasks: invalid input: "
	if i := strings.Index(s, mark); i >= 0 {
		return strings.TrimSpace(s[i+len(mark):])
	}
	return s
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func conflictDetail(err error) string {
	s := err.Error()
	const mark = "tasks: conflict: "
	if i := strings.Index(s, mark); i >= 0 {
		return strings.TrimSpace(s[i+len(mark):])
	}
	return s
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("%w: id", domain.ErrInvalidInput)
	}
	if len(id) > maxPathIDBytes {
		return "", fmt.Errorf("%w: id too long", domain.ErrInvalidInput)
	}
	return id, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathItemID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("%w: item id", domain.ErrInvalidInput)
	}
	if len(id) > maxPathIDBytes {
		return "", fmt.Errorf("%w: item id too long", domain.ErrInvalidInput)
	}
	return id, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func actorFromRequest(r *http.Request) domain.Actor {
	if r == nil {
		return domain.ActorUser
	}
	switch strings.ToLower(strings.TrimSpace(r.Header.Get("X-Actor"))) {
	case "agent":
		return domain.ActorAgent
	default:
		return domain.ActorUser
	}
}

func requestCtx(r *http.Request) context.Context {
	if r == nil {
		return context.Background()
	}
	return r.Context()
}

func requestRouteLabel(r *http.Request) string {
	if r == nil {
		return ""
	}
	if r.Pattern != "" {
		return r.Pattern
	}
	if r.URL != nil {
		return r.URL.Path
	}
	return ""
}

func logResponseWriteFailure(ctx context.Context, r *http.Request, op string, err error, stage string) {
	rid := ""
	route := ""
	if r != nil {
		rid = logctx.RequestIDFromContext(ctx)
		route = requestRouteLabel(r)
	}
	slog.Log(ctx, slog.LevelError, "response write failed",
		"cmd", calltrace.LogCmd, "operation", op,
		"request_id", rid, "route", route,
		"failure_stage", stage, "err", err)
}

func debugHTTPRequest(r *http.Request, op string, extra ...any) {
	if r == nil || !slog.Default().Enabled(r.Context(), slog.LevelDebug) {
		return
	}
	q := r.URL.RawQuery
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

func debugHTTPOut(ctx context.Context, op string, httpStatus int, extra ...any) {
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

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func truncateRunes(s string, maxRunes int) string {
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
