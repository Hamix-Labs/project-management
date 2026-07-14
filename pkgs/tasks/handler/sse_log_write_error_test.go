package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

func TestLogSSEWriteError_logsWhenClientContextActive(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })
	slog.SetDefault(slog.New(logctx.WrapSlogHandlerWithRequestContext(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))))

	r := httptest.NewRequest(http.MethodGet, "/events", nil)
	r = r.WithContext(logctx.ContextWithRequestID(r.Context(), "sse-rid"))
	logSSEWriteError(r, "tasks.sse", errors.New("simulated write failure"))

	var line map[string]any
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatal(err)
	}
	if line["msg"] != "sse write failed" {
		t.Fatalf("msg %v", line["msg"])
	}
	if line["request_id"] != "sse-rid" {
		t.Fatalf("request_id %v", line["request_id"])
	}
}

func TestLogSSEWriteError_skipsWhenRequestContextCanceled(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := httptest.NewRequest(http.MethodGet, "/events", nil)
	r = r.WithContext(ctx)
	logSSEWriteError(r, "tasks.sse", errors.New("would log if not canceled"))

	if strings.TrimSpace(buf.String()) != "" {
		t.Fatalf("expected no log, got %q", buf.String())
	}
}
