package handlerhttp

import (
	"fmt"
	"log/slog"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// MaxPathIDBytes caps path segments for task UUIDs, draft ids, checklist item
// ids, project ids, and similar route parameters.
const MaxPathIDBytes = 128

// ParsePathID trims raw and rejects empty or overlong values. The ErrInvalidInput
// detail uses the field name "id".
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ParsePathID(raw string) (string, error) {
	return ParseBoundedPathID("handlerhttp.ParsePathID", raw, "id")
}

// ParseBoundedPathID trims raw and rejects empty or overlong values.
// op is the debug-trace operation name; field names the ErrInvalidInput detail
// (e.g. "id", "item id", "cycle id").
func ParseBoundedPathID(op, raw, field string) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("%w: %s", taskcoredomain.ErrInvalidInput, field)
	}
	if len(id) > MaxPathIDBytes {
		return "", fmt.Errorf("%w: %s too long", taskcoredomain.ErrInvalidInput, field)
	}
	return id, nil
}
