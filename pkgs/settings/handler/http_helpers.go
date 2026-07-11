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

	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

const maxHTTPLogQueryBytes = 1024

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
	return fmt.Errorf("%w: json trailing data", taskcoredomain.ErrInvalidInput)
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

//funclogmeasure:skip category=delegate-already-logs reason="JSON response helper; HTTP handler chokepoint emits trace."
func writeJSONWithETag(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	apijson.ApplyRevalidatableHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		writeJSONError(w, r, op, http.StatusInternalServerError, "internal server error")
		return
	}
	payload := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
	etag := apijson.ComputeETag(payload)
	w.Header().Set("ETag", etag)
	if r != nil && apijson.IfNoneMatchMatches(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.WriteHeader(code)
	_, _ = w.Write(payload)
	_, _ = w.Write([]byte("\n"))
}

//funclogmeasure:skip category=delegate-already-logs reason="Error response helper; HTTP handler chokepoint emits trace."
func writeJSONError(w http.ResponseWriter, r *http.Request, op string, code int, msg string) {
	apijson.WriteJSONError(w, r, op, code, msg, calltrace.Path)
}

//funclogmeasure:skip category=delegate-already-logs reason="Error response helper; HTTP handler chokepoint emits trace."
func writeError(w http.ResponseWriter, r *http.Request, op string, err error, code int) {
	msg := http.StatusText(code)
	if code == http.StatusBadRequest {
		msg = err.Error()
	}
	writeJSONError(w, r, op, code, msg)
}

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
	writeJSONError(w, r, op, code, msg)
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
