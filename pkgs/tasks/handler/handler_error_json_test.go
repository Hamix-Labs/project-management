package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

func TestWriteJSONError_includes_request_id_from_context(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(logctx.ContextWithRequestID(req.Context(), "unit-rid-1"))
	apijson.WriteJSONError(rec, req, "test.op", http.StatusBadRequest, "bad", calltrace.Path)
	var out struct {
		Error     string `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Error != "bad" || out.RequestID != "unit-rid-1" {
		t.Fatalf("got %+v", out)
	}
}
