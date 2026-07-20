package handlerhttp

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

	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	gitinventoryhandler "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/handler"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

const maxHTTPLogJSONPreviewBytes = 16384

// ActorFromRequest reads X-Actor (user default, agent when set).
func ActorFromRequest(r *http.Request) taskcoredomain.Actor {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.ActorFromRequest")
	if r == nil {
		return taskcoredomain.ActorUser
	}
	ctx := calltrace.Push(r.Context(), "actorFromRequest")
	raw := strings.TrimSpace(r.Header.Get("X-Actor"))
	calltrace.HelperIOIn(ctx, "actorFromRequest", "x_actor_raw", raw)
	var a taskcoredomain.Actor
	defer func() {
		calltrace.HelperIOOut(ctx, "actorFromRequest", "actor", string(a))
	}()
	switch strings.ToLower(raw) {
	case "agent":
		a = taskcoredomain.ActorAgent
	default:
		a = taskcoredomain.ActorUser
	}
	return a
}

// DecodeJSON decodes a single JSON value from r into dst; rejects trailing data.
func DecodeJSON(ctx context.Context, r io.Reader, dst any) (err error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.DecodeJSON")
	ctx = calltrace.Push(ctx, "decodeJSON")
	calltrace.HelperIOIn(ctx, "decodeJSON", "dst_type", fmt.Sprintf("%T", dst))
	defer func() { calltrace.HelperIOOut(ctx, "decodeJSON", "err", err) }()
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err = dec.Decode(dst); err != nil {
		err = fmt.Errorf("json decode: %w", err)
		return err
	}
	if err = dec.Decode(&struct{}{}); err != nil {
		if err == io.EOF {
			err = nil
			return nil
		}
		err = fmt.Errorf("json trailing data: %w", err)
		return err
	}
	err = fmt.Errorf("%w: json trailing data", taskcoredomain.ErrInvalidInput)
	return err
}

// SetAPISecurityHeaders sets baseline hardening headers for browser-facing responses.
func SetAPISecurityHeaders(w http.ResponseWriter) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.SetAPISecurityHeaders")
	apijson.ApplySecurityHeaders(w)
}

// WrapPrometheusHandler applies baseline security headers before Prometheus scrape.
func WrapPrometheusHandler(next http.Handler) http.Handler {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.WrapPrometheusHandler")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apijson.ApplySecurityHeaders(w)
		next.ServeHTTP(w, r)
	})
}

func setJSONHeaders(w http.ResponseWriter) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.setJSONHeaders")
	apijson.ApplySecurityHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// WriteJSON writes v as JSON with optional debug http.io logging.
func WriteJSON(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	setJSONHeaders(w)
	ctx := RequestCtx(r)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		if r != nil {
			rid := logctx.RequestIDFromContext(ctx)
			route := RequestRouteLabel(r)
			slog.Log(ctx, slog.LevelError, "response encode failed",
				"cmd", calltrace.LogCmd, "operation", op, "request_id", rid, "route", route,
				"failure_stage", "response_encode", "err", err)
		} else {
			slog.Error("response encode failed", "cmd", calltrace.LogCmd, "operation", op,
				"failure_stage", "response_encode", "err", err)
		}
		WriteJSONError(w, r, op, http.StatusInternalServerError, "internal server error")
		return
	}
	payload := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
	if r != nil && slog.Default().Enabled(ctx, slog.LevelDebug) {
		preview := apijson.TruncateUTF8ByBytes(string(payload), maxHTTPLogJSONPreviewBytes)
		slog.Log(ctx, slog.LevelDebug, "http.io",
			"cmd", calltrace.LogCmd, "obs_category", "http_io", "operation", op, "call_path", calltrace.Path(ctx), "phase", "out",
			"http_status", code, "response_json_bytes", len(payload), "response_body", preview)
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

// WriteJSONError writes a JSON error envelope.
func WriteJSONError(w http.ResponseWriter, r *http.Request, op string, code int, msg string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.WriteJSONError", "http_op", op, "http_status", code)
	apijson.WriteJSONError(w, r, op, code, msg, calltrace.Path)
}

// WriteJSONWithETag serves JSON with strong ETag and 304 support.
func WriteJSONWithETag(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	apijson.ApplyRevalidatableHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	ctx := RequestCtx(r)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		if r != nil {
			rid := logctx.RequestIDFromContext(ctx)
			route := RequestRouteLabel(r)
			slog.Log(ctx, slog.LevelError, "response encode failed",
				"cmd", calltrace.LogCmd, "operation", op, "request_id", rid, "route", route,
				"failure_stage", "response_encode", "err", err)
		} else {
			slog.Error("response encode failed", "cmd", calltrace.LogCmd, "operation", op,
				"failure_stage", "response_encode", "err", err)
		}
		WriteJSONError(w, r, op, http.StatusInternalServerError, "internal server error")
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

// UserFacingJSONError maps decode/validation errors to client messages.
func UserFacingJSONError(err error) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.UserFacingJSONError")
	s := err.Error()
	if strings.HasPrefix(s, "json decode: ") {
		return strings.TrimPrefix(s, "json decode: ")
	}
	if errors.Is(err, taskcoredomain.ErrInvalidInput) {
		return "request body must contain a single JSON value"
	}
	if strings.HasPrefix(s, "json trailing data:") {
		return "request body must contain a single JSON value"
	}
	return s
}

// StoreErrorClientMessage maps domain/store errors to JSON error text.
func StoreErrorClientMessage(err error) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.StoreErrorClientMessage")
	switch {
	case errors.Is(err, taskcoredomain.ErrNotFound), errors.Is(err, projectsdomain.ErrNotFound):
		return "not found"
	case errors.Is(err, taskcoredomain.ErrConflict), errors.Is(err, projectsdomain.ErrConflict):
		if d := conflictDetail(err); d != "" {
			return d
		}
		if errors.Is(err, projectsdomain.ErrConflict) {
			return "conflict"
		}
		return "task id already exists"
	case errors.Is(err, taskcoredomain.ErrInvalidInput),
		errors.Is(err, projectsdomain.ErrInvalidInput),
		errors.Is(err, settingsdomain.ErrInvalidInput):
		if d := InvalidInputDetail(err); d != "" {
			return d
		}
		return "bad request"
	default:
		return "internal server error"
	}
}

// InvalidInputDetail extracts the client-facing suffix after a known
// invalid-input mark (projects, settings, or tasks). Implementation lives in
// apijson so gitinventory can share it without importing this package
// (import-cycle exception).
func InvalidInputDetail(err error) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.InvalidInputDetail")
	return apijson.InvalidInputDetail(err,
		apijson.ProjectsInvalidInputMark,
		apijson.SettingsInvalidInputMark,
		apijson.TasksInvalidInputMark,
	)
}

func conflictDetail(err error) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.conflictDetail")
	s := err.Error()
	for _, mark := range []string{"projects: conflict: ", "tasks: conflict: "} {
		if i := strings.Index(s, mark); i >= 0 {
			return strings.TrimSpace(s[i+len(mark):])
		}
	}
	return ""
}

// WriteError writes a JSON error for non-store failures (decode, max body, etc.).
func WriteError(w http.ResponseWriter, r *http.Request, op string, err error, code int) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.WriteError", "http_op", op)
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		code = http.StatusRequestEntityTooLarge
	}
	ctxErr := calltrace.Push(RequestCtx(r), "writeError")
	calltrace.HelperIOIn(ctxErr, "writeError", "http_op", op, "http_status", code, "err", err)
	logRequestFailure(r, op, err, code)
	msg := http.StatusText(code)
	switch code {
	case http.StatusRequestEntityTooLarge:
		msg = "request body too large"
	case http.StatusBadRequest:
		msg = UserFacingJSONError(err)
		if msg == "" {
			msg = "bad request"
		}
	}
	calltrace.HelperIOOut(ctxErr, "writeError", "client_facing_msg", msg)
	WriteJSONError(w, r, op, code, msg)
}

// StoreErrHTTPResponse maps store/domain errors to HTTP status and client message.
func StoreErrHTTPResponse(ctx context.Context, err error) (code int, msg string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.StoreErrHTTPResponse")
	ctx = calltrace.Push(ctx, "storeErrHTTPResponse")
	calltrace.HelperIOIn(ctx, "storeErrHTTPResponse", "err", err)
	defer func() {
		calltrace.HelperIOOut(ctx, "storeErrHTTPResponse", "http_status", code, "client_msg", msg)
	}()
	code = http.StatusInternalServerError
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		code = http.StatusGatewayTimeout
		msg = "request timed out"
		return code, msg
	case errors.Is(err, context.Canceled):
		code = http.StatusRequestTimeout
		msg = "request canceled"
		return code, msg
	case errors.Is(err, taskcoredomain.ErrNotFound), errors.Is(err, projectsdomain.ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, taskcoredomain.ErrInvalidInput),
		errors.Is(err, projectsdomain.ErrInvalidInput),
		errors.Is(err, settingsdomain.ErrInvalidInput):
		code = http.StatusBadRequest
	case errors.Is(err, taskcoredomain.ErrConflict), errors.Is(err, projectsdomain.ErrConflict):
		code = http.StatusConflict
	}
	msg = StoreErrorClientMessage(err)
	if code == http.StatusInternalServerError {
		msg = "internal server error"
	}
	return code, msg
}

// WriteStoreError maps store/domain errors to HTTP status + JSON body.
func WriteStoreError(w http.ResponseWriter, r *http.Request, op string, err error, logExtras ...any) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.WriteStoreError", "http_op", op)
	if gitdomain.GitErrCode(err) != "" {
		gitinventoryhandler.WriteGitStoreError(w, r, op, err)
		return
	}
	ctxErr := calltrace.Push(RequestCtx(r), "writeStoreError")
	calltrace.HelperIOIn(ctxErr, "writeStoreError", "http_op", op, "err", err)
	code, msg := StoreErrHTTPResponse(ctxErr, err)
	calltrace.HelperIOOut(ctxErr, "writeStoreError", "http_status", code, "client_facing_msg", msg)
	logRequestFailure(r, op, err, code, logExtras...)
	WriteJSONError(w, r, op, code, msg)
}

// RequestCtx returns r.Context() or background when r is nil.
func RequestCtx(r *http.Request) context.Context {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.RequestCtx")
	if r == nil {
		return context.Background()
	}
	return r.Context()
}

// RequestRouteLabel returns the route pattern or URL path for logs.
func RequestRouteLabel(r *http.Request) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handlerhttp.RequestRouteLabel")
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

func logRequestFailure(r *http.Request, op string, err error, httpStatus int, extra ...any) {
	ctx := RequestCtx(r)
	rid := ""
	route := ""
	if r != nil {
		rid = logctx.RequestIDFromContext(ctx)
		route = RequestRouteLabel(r)
	}
	attrs := []any{
		"cmd", calltrace.LogCmd, "operation", op,
		"call_path", calltrace.Path(ctx),
		"http_status", httpStatus, "err", err,
		"request_id", rid, "route", route,
	}
	attrs = append(attrs, extra...)
	if httpStatus >= 500 {
		slog.Log(ctx, slog.LevelError, "request failed", attrs...)
		return
	}
	slog.Log(ctx, slog.LevelWarn, "request failed", attrs...)
}

func logResponseWriteFailure(ctx context.Context, r *http.Request, op string, err error, stage string) {
	rid := ""
	route := ""
	if r != nil {
		rid = logctx.RequestIDFromContext(ctx)
		route = RequestRouteLabel(r)
	}
	slog.Log(ctx, slog.LevelError, "response write failed",
		"cmd", calltrace.LogCmd, "operation", op,
		"request_id", rid, "route", route,
		"failure_stage", stage, "err", err)
}

// LogRequestFailure logs warn/error for failed requests (exported for handler package tests).
func LogRequestFailure(r *http.Request, op string, err error, httpStatus int, extra ...any) {
	logRequestFailure(r, op, err, httpStatus, extra...)
}
