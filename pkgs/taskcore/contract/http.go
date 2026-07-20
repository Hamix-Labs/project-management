package contract

import (
	"context"
	"io"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// HTTPHelpers is the narrow HTTP response/request port injected by composition.
// Implementations live in pkgs/tasks/handlerhttp; taskcore must not import that package.
type HTTPHelpers interface {
	ActorFromRequest(r *http.Request) domain.Actor
	DecodeJSON(ctx context.Context, body io.Reader, dst any) error
	WriteJSON(w http.ResponseWriter, r *http.Request, op string, code int, v any)
	WriteJSONWithETag(w http.ResponseWriter, r *http.Request, op string, code int, v any)
	WriteError(w http.ResponseWriter, r *http.Request, op string, err error, code int)
	WriteStoreError(w http.ResponseWriter, r *http.Request, op string, err error, logExtras ...any)
	ParseBoundedPathID(op, raw, field string) (string, error)
	DebugHTTPRequest(r *http.Request, op string, extra ...any)
	DebugHTTPOut(ctx context.Context, op string, httpStatus int, extra ...any)
}
