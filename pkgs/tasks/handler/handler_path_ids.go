package handler

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

// maxTaskPathIDBytes caps path segments for task UUIDs, draft ids, checklist item ids, etc.
const maxTaskPathIDBytes = 128

// maxPhaseSeqParamBytes caps the {phaseSeq} path segment so strconv work and
// log fields stay bounded. Phase sequences are small positive integers; the
// 32-byte cap matches the documented event-seq cap.
const maxPhaseSeqParamBytes = 32

func parseBoundedPathID(op, raw, field string) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", op)
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("%w: %s", taskcoredomain.ErrInvalidInput, field)
	}
	if len(id) > maxTaskPathIDBytes {
		return "", fmt.Errorf("%w: %s too long", taskcoredomain.ErrInvalidInput, field)
	}
	return id, nil
}

func parseTaskPathID(id string) (string, error) {
	return parseBoundedPathID("handler.parseTaskPathID", id, "id")
}

func parseTaskPathItemID(id string) (string, error) {
	return parseBoundedPathID("handler.parseTaskPathItemID", id, "item id")
}

// parseTaskPathCycleID validates the {cycleId} path segment for the cycles
// resource family (same UUID-shape and 128-byte cap as task ids; the bare
// field name "cycle id" surfaces in the 400 message).
func parseTaskPathCycleID(id string) (string, error) {
	return parseBoundedPathID("handler.parseTaskPathCycleID", id, "cycle id")
}

// parseTaskPathPhaseSeq validates the {phaseSeq} path segment. Sequences
// are positive int64 values; whitespace, non-numeric, zero, and negative
// values are rejected.
func parseTaskPathPhaseSeq(raw string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.parseTaskPathPhaseSeq")
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
