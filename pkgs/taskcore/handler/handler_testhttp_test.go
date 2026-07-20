package handler_test

import (
	"context"
	"io"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

// testHTTP adapts handlerhttp for unit tests that Register taskcore routes
// without the composition shell.
type testHTTP struct{}

func (testHTTP) ActorFromRequest(r *http.Request) domain.Actor {
	return handlerhttp.ActorFromRequest(r)
}

func (testHTTP) DecodeJSON(ctx context.Context, body io.Reader, dst any) error {
	return handlerhttp.DecodeJSON(ctx, body, dst)
}

func (testHTTP) WriteJSON(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	handlerhttp.WriteJSON(w, r, op, code, v)
}

func (testHTTP) WriteJSONWithETag(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	handlerhttp.WriteJSONWithETag(w, r, op, code, v)
}

func (testHTTP) WriteError(w http.ResponseWriter, r *http.Request, op string, err error, code int) {
	handlerhttp.WriteError(w, r, op, err, code)
}

func (testHTTP) WriteStoreError(w http.ResponseWriter, r *http.Request, op string, err error, logExtras ...any) {
	handlerhttp.WriteStoreError(w, r, op, err, logExtras...)
}

func (testHTTP) ParseBoundedPathID(op, raw, field string) (string, error) {
	return handlerhttp.ParseBoundedPathID(op, raw, field)
}

func (testHTTP) DebugHTTPRequest(r *http.Request, op string, extra ...any) {
	handlerhttp.DebugHTTPRequest(r, op, extra...)
}

func (testHTTP) DebugHTTPOut(ctx context.Context, op string, httpStatus int, extra ...any) {
	handlerhttp.DebugHTTPOut(ctx, op, httpStatus, extra...)
}

func testDeps(tasks contract.TaskCRUDStore) handler.Deps {
	return handler.Deps{Tasks: tasks, HTTP: testHTTP{}}
}
