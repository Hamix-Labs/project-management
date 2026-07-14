package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

func captureHandlerLogs(t *testing.T, level slog.Level) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })
	var processSeq atomic.Uint64
	base := logctx.WrapSlogHandlerWithRequestContext(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(slog.New(logctx.WrapSlogHandlerWithLogSequence(base, &processSeq)))
	return &buf
}

func lastLogLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("want at least one log line, got %q", buf.String())
	}
	var line map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &line); err != nil {
		t.Fatalf("decode log: %v raw=%q", err, lines[len(lines)-1])
	}
	return line
}

type errAfterHeaderWriter struct {
	http.ResponseWriter
	headerWritten bool
}

func (w *errAfterHeaderWriter) WriteHeader(code int) {
	w.headerWritten = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *errAfterHeaderWriter) Write(b []byte) (int, error) {
	if w.headerWritten {
		return 0, errors.New("simulated response write failure")
	}
	return w.ResponseWriter.Write(b)
}

func TestWriteJSON_responseWriteFailure_logs(t *testing.T) {
	buf := captureHandlerLogs(t, slog.LevelError)

	inner := httptest.NewRecorder()
	w := &errAfterHeaderWriter{ResponseWriter: inner}
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req = req.WithContext(logctx.ContextWithRequestID(req.Context(), "list-write-rid"))

	handlerhttp.WriteJSON(w, req, "tasks.list", http.StatusOK, map[string]any{"tasks": []any{}})

	line := lastLogLine(t, buf)
	if line["msg"] != "response write failed" {
		t.Fatalf("msg=%v want response write failed", line["msg"])
	}
	if line["operation"] != "tasks.list" {
		t.Fatalf("operation=%v", line["operation"])
	}
	if line["failure_stage"] != "body" {
		t.Fatalf("failure_stage=%v want body", line["failure_stage"])
	}
	if line["request_id"] != "list-write-rid" {
		t.Fatalf("request_id=%v", line["request_id"])
	}
}

func TestWriteJSONWithETag_responseWriteFailure_logs(t *testing.T) {
	buf := captureHandlerLogs(t, slog.LevelError)

	inner := httptest.NewRecorder()
	w := &errAfterHeaderWriter{ResponseWriter: inner}
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req = req.WithContext(logctx.ContextWithRequestID(req.Context(), "list-etag-write-rid"))

	handlerhttp.WriteJSONWithETag(w, req, "tasks.list", http.StatusOK, map[string]any{"tasks": []any{}})

	line := lastLogLine(t, buf)
	if line["msg"] != "response write failed" {
		t.Fatalf("msg=%v want response write failed", line["msg"])
	}
	if line["operation"] != "tasks.list" {
		t.Fatalf("operation=%v", line["operation"])
	}
	if line["failure_stage"] != "body" {
		t.Fatalf("failure_stage=%v want body", line["failure_stage"])
	}
}

func TestWriteJSONWithETag_encodeFailure_logs(t *testing.T) {
	buf := captureHandlerLogs(t, slog.LevelError)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req = req.WithContext(logctx.ContextWithRequestID(req.Context(), "list-etag-encode-rid"))

	handlerhttp.WriteJSONWithETag(rr, req, "tasks.list", http.StatusOK, map[string]any{"bad": make(chan int)})

	line := lastLogLine(t, buf)
	if line["msg"] != "response encode failed" {
		t.Fatalf("msg=%v want response encode failed", line["msg"])
	}
	if line["failure_stage"] != "response_encode" {
		t.Fatalf("failure_stage=%v want response_encode", line["failure_stage"])
	}
	if line["operation"] != "tasks.list" {
		t.Fatalf("operation=%v", line["operation"])
	}
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status %d want 500", rr.Code)
	}
}
