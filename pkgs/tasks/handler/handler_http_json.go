package handler

import (
	"context"
	"io"
	"net/http"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

// jsonErrorBody is the standard API error envelope used in handler contract tests.
type jsonErrorBody struct {
	Error string `json:"error"`
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func actorFromRequest(r *http.Request) taskcoredomain.Actor {
	return handlerhttp.ActorFromRequest(r)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func decodeJSON(ctx context.Context, r io.Reader, dst any) error {
	return handlerhttp.DecodeJSON(ctx, r, dst)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func setAPISecurityHeaders(w http.ResponseWriter) {
	handlerhttp.SetAPISecurityHeaders(w)
}

// WrapPrometheusHandler re-exports the shared metrics wrapper for cmd/taskapi.
//
//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func WrapPrometheusHandler(next http.Handler) http.Handler {
	return handlerhttp.WrapPrometheusHandler(next)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func writeJSON(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	handlerhttp.WriteJSON(w, r, op, code, v)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func writeJSONError(w http.ResponseWriter, r *http.Request, op string, code int, msg string) {
	handlerhttp.WriteJSONError(w, r, op, code, msg)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func writeJSONWithETag(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	handlerhttp.WriteJSONWithETag(w, r, op, code, v)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func userFacingJSONError(err error) string {
	return handlerhttp.UserFacingJSONError(err)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func storeErrorClientMessage(err error) string {
	return handlerhttp.StoreErrorClientMessage(err)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func writeError(w http.ResponseWriter, r *http.Request, op string, err error, code int) {
	handlerhttp.WriteError(w, r, op, err, code)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func writeStoreError(w http.ResponseWriter, r *http.Request, op string, err error, logExtras ...any) {
	handlerhttp.WriteStoreError(w, r, op, err, logExtras...)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func requestCtx(r *http.Request) context.Context {
	return handlerhttp.RequestCtx(r)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func requestRouteLabel(r *http.Request) string {
	return handlerhttp.RequestRouteLabel(r)
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin forwarder to pkgs/tasks/handlerhttp."
func logRequestFailure(r *http.Request, op string, err error, httpStatus int, extra ...any) {
	handlerhttp.LogRequestFailure(r, op, err, httpStatus, extra...)
}
