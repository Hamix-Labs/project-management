package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// startContractServer returns a test server and registers Close on t.Cleanup.
func startContractServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := newTaskTestServer(t)
	t.Cleanup(srv.Close)
	return srv
}

// mustGetJSON GETs path and fatals unless status is 200.
func mustGetJSON(t *testing.T, baseURL, path string) ([]byte, *http.Response) {
	t.Helper()
	res, err := http.Get(baseURL + path)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d (want 200) body=%s", path, res.StatusCode, raw)
	}
	return raw, res
}

// assertBareError asserts status and exact jsonErrorBody.error string.
func assertBareError(t *testing.T, res *http.Response, raw []byte, wantStatus int, wantError string) {
	t.Helper()
	if res.StatusCode != wantStatus {
		t.Fatalf("status %d (want %d) body=%s", res.StatusCode, wantStatus, raw)
	}
	var errBody jsonErrorBody
	if err := json.Unmarshal(raw, &errBody); err != nil {
		t.Fatalf("decode: %v body=%s", err, raw)
	}
	if errBody.Error != wantError {
		t.Fatalf("error=%q want %q", errBody.Error, wantError)
	}
}

// assertHTTPError asserts status and that jsonErrorBody.error contains substr.
func assertHTTPError(t *testing.T, res *http.Response, raw []byte, wantStatus int, substr string) {
	t.Helper()
	if res.StatusCode != wantStatus {
		t.Fatalf("status %d (want %d) body=%s", res.StatusCode, wantStatus, raw)
	}
	var errBody jsonErrorBody
	if err := json.Unmarshal(raw, &errBody); err != nil {
		t.Fatalf("decode: %v body=%s", err, raw)
	}
	if !strings.Contains(errBody.Error, substr) {
		t.Fatalf("error=%q want substring %q", errBody.Error, substr)
	}
}

// equalStringSlices reports whether a and b are the same length with equal elements in order.
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
