package shell_test

import (
	"encoding/json"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/middleware"
)

func TestHTTP_error_JSON_includes_request_id_with_access_middleware(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	api := middleware.WithAccessLog(handler.NewHandler(composition.NewAPI(db), realtime.NewSSEHub(), nil), calltrace.Path)
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/tasks", "application/json", strings.NewReader(`{"unknown_field":true}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
	ridHeader := strings.TrimSpace(res.Header.Get("X-Request-ID"))
	if ridHeader == "" {
		t.Fatal("missing X-Request-ID on response")
	}
	var out struct {
		Error     string `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	const wantErr = `json: unknown field "unknown_field"`
	if out.Error != wantErr {
		t.Fatalf("error %q want %q (encoding/json with DisallowUnknownFields)", out.Error, wantErr)
	}
	if out.RequestID != ridHeader {
		t.Fatalf("request_id %q want %q", out.RequestID, ridHeader)
	}
}
