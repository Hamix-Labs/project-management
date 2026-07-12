package handler

import (
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

// WrapPrometheusHandler re-exports the shared metrics wrapper for cmd/taskapi.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to handlerhttp; metrics scrape traces at HTTP boundary."
func WrapPrometheusHandler(next http.Handler) http.Handler {
	return handlerhttp.WrapPrometheusHandler(next)
}

// jsonErrorBody is the standard API error envelope used in handler contract tests.
type jsonErrorBody struct {
	Error string `json:"error"`
}
