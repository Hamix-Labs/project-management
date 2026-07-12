package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

const (
	maxPathIDBytes      = 128
	maxHTTPLogTextRunes = 240
)

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
func parseTaskPathItemID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("%w: item id", taskcoredomain.ErrInvalidInput)
	}
	if len(id) > maxPathIDBytes {
		return "", fmt.Errorf("%w: item id too long", taskcoredomain.ErrInvalidInput)
	}
	return id, nil
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
