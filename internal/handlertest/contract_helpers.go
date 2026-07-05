package handlertest

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// AssertJSONKeys fails when raw JSON is missing any of wantKeys (contract shape guard).
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only contract helper; not part of production trace paths."
func AssertJSONKeys(t *testing.T, raw []byte, wantKeys ...string) {
	t.Helper()
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatalf("decode: %v body=%s", err, raw)
	}
	for _, k := range wantKeys {
		if _, ok := top[k]; !ok {
			t.Errorf("response missing key %q: %s", k, raw)
		}
	}
}

// PostJSON issues POST with Content-Type application/json and returns the response body.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only contract helper; not part of production trace paths."
func PostJSON(t *testing.T, baseURL, path, body string) (*http.Response, []byte) {
	t.Helper()
	res, err := http.Post(baseURL+path, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res, raw
}
