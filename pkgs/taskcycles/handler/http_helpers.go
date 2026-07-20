package handler

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

const (
	maxPhaseSeqParamBytes     = 32
	maxListIntQueryParamBytes = 32
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathID(id string) (string, error) {
	return handlerhttp.ParseBoundedPathID("taskcycles.handler.parseTaskPathID", id, "id")
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathCycleID(id string) (string, error) {
	return handlerhttp.ParseBoundedPathID("taskcycles.handler.parseTaskPathCycleID", id, "cycle id")
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
