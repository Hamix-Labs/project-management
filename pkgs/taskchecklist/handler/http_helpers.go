package handler

import (
	"context"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

const maxHTTPLogTextRunes = 240

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathID(id string) (string, error) {
	return handlerhttp.ParseBoundedPathID("taskchecklist.handler.parseTaskPathID", id, "id")
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathItemID(id string) (string, error) {
	return handlerhttp.ParseBoundedPathID("taskchecklist.handler.parseTaskPathItemID", id, "item id")
}

//funclogmeasure:skip category=hot-path reason="Thin re-export of handlerhttp.DebugHTTPRequest; shared package emits the http.io trace."
func debugHTTPRequest(r *http.Request, op string, extra ...any) {
	handlerhttp.DebugHTTPRequest(r, op, extra...)
}

//funclogmeasure:skip category=hot-path reason="Thin re-export of handlerhttp.DebugHTTPOut; shared package emits the http.io trace."
func debugHTTPOut(ctx context.Context, op string, httpStatus int, extra ...any) {
	handlerhttp.DebugHTTPOut(ctx, op, httpStatus, extra...)
}

//funclogmeasure:skip category=hot-path reason="Thin re-export of handlerhttp.TruncateRunes; operation trace is emitted by the calling chokepoint."
func truncateRunes(s string, maxRunes int) string {
	return handlerhttp.TruncateRunes(s, maxRunes)
}
