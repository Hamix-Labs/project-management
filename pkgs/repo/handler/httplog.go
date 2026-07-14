package handler

import (
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

const maxHTTPLogTitleRunes = 160

//funclogmeasure:skip category=hot-path reason="Thin re-export of handlerhttp.DebugHTTPRequest; shared package emits the http.io trace."
func debugHTTPRequest(r *http.Request, op string, extra ...any) {
	handlerhttp.DebugHTTPRequest(r, op, extra...)
}

//funclogmeasure:skip category=hot-path reason="Thin re-export of handlerhttp.TruncateRunes; operation trace is emitted by the calling chokepoint."
func truncateRunes(s string, maxRunes int) string {
	return handlerhttp.TruncateRunes(s, maxRunes)
}
