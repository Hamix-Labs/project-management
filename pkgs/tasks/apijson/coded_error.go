package apijson

import (
	"encoding/json"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

// JSONCodedErrorBody is the wire shape for errors that include a machine code
// (git inventory conflicts/not-found, etc.).
type JSONCodedErrorBody struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// WriteJSONCodedError writes {"error","code","request_id"} with security headers.
//
//funclogmeasure:skip category=hot-path reason="JSON helper; HTTP handlers emit operation traces."
func WriteJSONCodedError(w http.ResponseWriter, r *http.Request, op string, status int, code, msg string) {
	_ = op // reserved for future http.io debug parity with WriteJSONError
	ApplySecurityHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	body := JSONCodedErrorBody{Error: msg, Code: code}
	if r != nil {
		body.RequestID = logctx.RequestIDFromContext(r.Context())
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(body)
}
