package desktop

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/AlexsanderHamir/Hamix/internal/desktopconfig"
)

func TestIsAPIPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/tasks", true},
		{"/tasks/abc", true},
		{"/events", true},
		{"/settings", true},
		{"/v1/bootstrap", true},
		{"/", false},
		{"/setup", false},
		{"/projects/x", true},
	}
	for _, tc := range cases {
		if got := isAPIPath(tc.path); got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.path, got, tc.want)
		}
	}
}

func TestSPAFallback_servesIndexForHTML(t *testing.T) {
	assets := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>ok</html>")},
	}
	h := NewHost(desktopconfig.WithRoot(t.TempDir()), assets)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	req.Header.Set("Accept", "text/html")
	h.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "ok") {
		t.Fatalf("body %q", rr.Body.String())
	}
}

func TestSPAFallback_apiWithoutRuntime_503(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	h := NewHost(desktopconfig.WithRoot(t.TempDir()), fstest.MapFS{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Accept", "application/json")
	h.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("code %d body %q", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "database not configured") {
		t.Fatalf("body %q", rr.Body.String())
	}
}

func TestSPAFallback_apiWithoutRuntime_dsnConfigured_honestError(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example.invalid/hamix")
	h := NewHost(desktopconfig.WithRoot(t.TempDir()), fstest.MapFS{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Accept", "application/json")
	h.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("code %d body %q", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if strings.Contains(body, "database not configured") {
		t.Fatalf("must not mask runtime failure as not configured: %q", body)
	}
	if !strings.Contains(body, "api runtime unavailable") {
		t.Fatalf("body %q", body)
	}
}
