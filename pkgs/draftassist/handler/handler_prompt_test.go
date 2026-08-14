package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
	draftassisthandler "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/handler"
	draftassiststore "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/store"
)

func newPromptHarness(t *testing.T) (*httptest.Server, string, string) {
	t.Helper()
	store := draftassiststore.NewMemoryStore()
	mux := http.NewServeMux()
	draftassisthandler.Register(mux, draftassisthandler.Deps{Store: store})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	sess, err := store.CreateSession(context.Background(), contract.CreateSessionInput{
		Snapshot: domain.FormSnapshot{Prompt: "<p>seed</p>"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv, sess.ID, sess.Nonce
}

func patchPrompt(t *testing.T, srv *httptest.Server, id, nonce, prompt string) *http.Response {
	t.Helper()
	body := strings.NewReader(`{"prompt":` + jsonString(prompt) + `}`)
	req, err := http.NewRequest(http.MethodPatch, srv.URL+"/draft-assist/sessions/"+id+"/prompt", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if nonce != "" {
		req.Header.Set(draftassisthandler.NonceHeader, nonce)
	}
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func TestPatchPrompt_missingNonce401(t *testing.T) {
	srv, id, _ := newPromptHarness(t)
	res := patchPrompt(t, srv, id, "", "<p>x</p>")
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status=%d", res.StatusCode)
	}
}

func TestPatchPrompt_wrongNonce403(t *testing.T) {
	srv, id, _ := newPromptHarness(t)
	res := patchPrompt(t, srv, id, "wrong", "<p>x</p>")
	defer res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("status=%d", res.StatusCode)
	}
}

func TestPatchPrompt_scriptRejected400(t *testing.T) {
	srv, id, nonce := newPromptHarness(t)
	res := patchPrompt(t, srv, id, nonce, "<p><script>alert(1)</script></p>")
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status=%d body=%s", res.StatusCode, b)
	}
}

func TestPatchPrompt_ok(t *testing.T) {
	srv, id, nonce := newPromptHarness(t)
	res := patchPrompt(t, srv, id, nonce, "<h2>Section</h2><p>ok</p>")
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status=%d body=%s", res.StatusCode, b)
	}
	var body struct {
		Snapshot struct {
			Prompt string `json:"prompt"`
		} `json:"snapshot"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body.Snapshot.Prompt, "Section") {
		t.Fatalf("prompt not updated: %+v", body)
	}
}
