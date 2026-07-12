package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/httpsecurityexpect"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
)

func TestStreamEvents_sets_security_headers(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	h := &Handler{store: composition.NewAPI(db), hub: NewSSEHub(), repoProv: NewStaticRepoProvider(nil)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	h.streamEvents(rec, req)
	httpsecurityexpect.AssertBaselineHeaders(t, rec.Header())
	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("Content-Type = %q", rec.Header().Get("Content-Type"))
	}
}
