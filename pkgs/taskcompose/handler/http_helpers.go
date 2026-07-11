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
	"strconv"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

const (
	maxPathIDBytes            = 128
	maxListIntQueryParamBytes = 32
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
	if errors.Is(err, taskcoredomain.ErrInvalidInput) {
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
	case errors.Is(err, taskcoredomain.ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, taskcoredomain.ErrInvalidInput):
		return http.StatusBadRequest, invalidInputDetail(err)
	case errors.Is(err, taskcoredomain.ErrConflict):
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
func parseBoundedLimit(q map[string][]string, def, max int) (int, error) {
	raw := strings.TrimSpace(firstQueryValue(q, "limit"))
	if raw == "" {
		return def, nil
	}
	if len(raw) > maxListIntQueryParamBytes {
		return 0, fmt.Errorf("%w: limit value too long", taskcoredomain.ErrInvalidInput)
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 || n > max {
		return 0, fmt.Errorf("%w: limit must be integer 0..%d", taskcoredomain.ErrInvalidInput, max)
	}
	if n == 0 {
		return def, nil
	}
	return n, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func firstQueryValue(q map[string][]string, key string) string {
	values := q[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	if len(id) > maxPathIDBytes {
		return "", fmt.Errorf("%w: id too long", taskcoredomain.ErrInvalidInput)
	}
	return id, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func actorFromRequest(r *http.Request) taskcoredomain.Actor {
	if r == nil {
		return taskcoredomain.ActorUser
	}
	switch strings.ToLower(strings.TrimSpace(r.Header.Get("X-Actor"))) {
	case "agent":
		return taskcoredomain.ActorAgent
	default:
		return taskcoredomain.ActorUser
	}
}
