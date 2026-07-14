package handler

import (
	"fmt"
	"net/http"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

const (
	maxPathIDBytes            = 128
	maxHTTPLogTextRunes       = 240
	maxTaskEventSeqParamBytes = 32
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

//funclogmeasure:skip category=hot-path reason="Thin re-export of handlerhttp.DebugHTTPRequest; shared package emits the http.io trace."
func debugHTTPRequest(r *http.Request, op string, extra ...any) {
	handlerhttp.DebugHTTPRequest(r, op, extra...)
}

//funclogmeasure:skip category=hot-path reason="Thin re-export of handlerhttp.TruncateRunes; operation trace is emitted by the calling chokepoint."
func truncateRunes(s string, maxRunes int) string {
	return handlerhttp.TruncateRunes(s, maxRunes)
}
