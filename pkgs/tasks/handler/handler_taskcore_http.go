package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

// taskcoreHTTP adapts pkgs/tasks/handlerhttp to taskcore/contract.HTTPHelpers.
type taskcoreHTTP struct{}

func (taskcoreHTTP) ActorFromRequest(r *http.Request) domain.Actor {
	return handlerhttp.ActorFromRequest(r)
}

func (taskcoreHTTP) DecodeJSON(ctx context.Context, body io.Reader, dst any) error {
	return handlerhttp.DecodeJSON(ctx, body, dst)
}

func (taskcoreHTTP) WriteJSON(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	handlerhttp.WriteJSON(w, r, op, code, v)
}

func (taskcoreHTTP) WriteJSONWithETag(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	handlerhttp.WriteJSONWithETag(w, r, op, code, v)
}

func (taskcoreHTTP) WriteError(w http.ResponseWriter, r *http.Request, op string, err error, code int) {
	handlerhttp.WriteError(w, r, op, err, code)
}

func (taskcoreHTTP) WriteStoreError(w http.ResponseWriter, r *http.Request, op string, err error, logExtras ...any) {
	handlerhttp.WriteStoreError(w, r, op, err, logExtras...)
}

func (taskcoreHTTP) ParseBoundedPathID(op, raw, field string) (string, error) {
	return handlerhttp.ParseBoundedPathID(op, raw, field)
}

func (taskcoreHTTP) DebugHTTPRequest(r *http.Request, op string, extra ...any) {
	handlerhttp.DebugHTTPRequest(r, op, extra...)
}

func (taskcoreHTTP) DebugHTTPOut(ctx context.Context, op string, httpStatus int, extra ...any) {
	handlerhttp.DebugHTTPOut(ctx, op, httpStatus, extra...)
}
