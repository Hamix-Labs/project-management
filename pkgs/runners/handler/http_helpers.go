package handler

import (
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

//funclogmeasure:skip category=hot-path reason="Thin re-export of handlerhttp.DebugHTTPRequest; shared package emits the http.io trace."
func debugHTTPRequest(r *http.Request, op string, extra ...any) {
	handlerhttp.DebugHTTPRequest(r, op, extra...)
}
