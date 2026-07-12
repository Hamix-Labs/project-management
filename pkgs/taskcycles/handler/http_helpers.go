package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

const (
	maxPathIDBytes            = 128
	maxPhaseSeqParamBytes     = 32
	maxListIntQueryParamBytes = 32
	maxHTTPLogTextRunes       = 240
)

func parseBoundedPathID(op, raw, field string) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("%w: %s", taskcoredomain.ErrInvalidInput, field)
	}
	if len(id) > maxPathIDBytes {
		return "", fmt.Errorf("%w: %s too long", taskcoredomain.ErrInvalidInput, field)
	}
	return id, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathID(id string) (string, error) {
	return parseBoundedPathID("taskcycles.handler.parseTaskPathID", id, "id")
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathCycleID(id string) (string, error) {
	return parseBoundedPathID("taskcycles.handler.parseTaskPathCycleID", id, "cycle id")
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathPhaseSeq(raw string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.parseTaskPathPhaseSeq")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("%w: phase_seq must be a positive integer", taskcoredomain.ErrInvalidInput)
	}
	if len(raw) > maxPhaseSeqParamBytes {
		return 0, fmt.Errorf("%w: phase_seq too long", taskcoredomain.ErrInvalidInput)
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("%w: phase_seq must be a positive integer", taskcoredomain.ErrInvalidInput)
	}
	return n, nil
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
