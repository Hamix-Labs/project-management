package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
	draftmcp "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/mcp"
	draftassiststore "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/store"
)

func TestUpdatePrompt_nonceFailClosed(t *testing.T) {
	store := draftassiststore.NewMemoryStore()
	sess, err := store.CreateSession(context.Background(), contract.CreateSessionInput{
		Snapshot: domain.FormSnapshot{Prompt: "<p>old</p>"},
	})
	if err != nil {
		t.Fatal(err)
	}
	host := &draftmcp.ToolHost{
		Bind:  &draftmcp.BindFile{SessionID: sess.ID, Nonce: sess.Nonce},
		Store: store,
	}
	if _, err := host.Store.UpdatePrompt(context.Background(), host.Bind.SessionID, "bad", "<p>new</p>"); err == nil {
		t.Fatal("expected nonce mismatch")
	}
	if _, err := host.Store.UpdatePrompt(context.Background(), host.Bind.SessionID, host.Bind.Nonce, "<p>new</p>"); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetSession(context.Background(), sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Snapshot.Prompt != "<p>new</p>" {
		t.Fatalf("prompt=%q", got.Snapshot.Prompt)
	}
}

// fakeTaskAPI is a minimal in-memory taskapi that supports the endpoints the
// draft-assist MCP HTTP client uses. It stores one session and exposes hooks
// tests use to drive read tools deterministically.
type fakeTaskAPI struct {
	sessionID  string
	nonce      string
	worktreeID string
	prompt     string
	files      []string
	fileBody   string
	templates  []map[string]any
	tasks      []map[string]any
	patchCalls int
}

func newFakeTaskAPI() *fakeTaskAPI {
	return &fakeTaskAPI{
		sessionID:  "sess-1",
		nonce:      "nonce-1",
		worktreeID: "wt-1",
		prompt:     "<p>hello</p>",
		files:      []string{"pkgs/draftassist/README.md", "cmd/hamix-draft-mcp/main.go"},
		fileBody:   "line1\nline2\n",
		templates:  []map[string]any{{"id": "tmpl-1", "name": "Bug template"}},
		tasks: []map[string]any{
			{"id": "t-1", "title": "Refactor draft-assist"},
			{"id": "t-2", "title": "Fix login bug"},
		},
	}
}

func (f *fakeTaskAPI) handler(t *testing.T) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /draft-assist/sessions/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/draft-assist/sessions/")
		if id != f.sessionID {
			http.Error(w, "no session", http.StatusNotFound)
			return
		}
		writeJSON(w, map[string]any{
			"id":          f.sessionID,
			"nonce":       f.nonce,
			"worktree_id": f.worktreeID,
			"snapshot":    map[string]any{"prompt": f.prompt, "title": "compose"},
		})
	})
	mux.HandleFunc("PATCH /draft-assist/sessions/{id}/prompt", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id != f.sessionID {
			http.Error(w, "no session", http.StatusNotFound)
			return
		}
		got := strings.TrimSpace(r.Header.Get(draftmcp.NonceHeader))
		if got == "" {
			http.Error(w, "nonce required", http.StatusUnauthorized)
			return
		}
		if got != f.nonce {
			http.Error(w, "nonce mismatch", http.StatusForbidden)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Prompt string `json:"prompt"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if err := domain.ValidateHTML(payload.Prompt); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.prompt = payload.Prompt
		f.patchCalls++
		writeJSON(w, map[string]any{"id": f.sessionID, "snapshot": map[string]any{"prompt": f.prompt}})
	})
	mux.HandleFunc("GET /repo/files", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("worktree_id") != f.worktreeID {
			http.Error(w, "worktree not found", http.StatusNotFound)
			return
		}
		q := strings.ToLower(r.URL.Query().Get("q"))
		matches := make([]string, 0)
		for _, p := range f.files {
			if q == "" || strings.Contains(strings.ToLower(p), q) {
				matches = append(matches, p)
			}
		}
		writeJSON(w, map[string]any{"paths": matches, "has_more": false, "truncated": false})
	})
	mux.HandleFunc("GET /repo/file", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("worktree_id") != f.worktreeID {
			http.Error(w, "worktree not found", http.StatusNotFound)
			return
		}
		p := r.URL.Query().Get("path")
		writeJSON(w, map[string]any{
			"path": p, "content": f.fileBody, "binary": false, "truncated": false,
			"size_bytes": int64(len(f.fileBody)), "line_count": 2,
		})
	})
	mux.HandleFunc("GET /task-templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"templates": f.templates})
	})
	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"tasks": f.tasks})
	})
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func newFakeServer(t *testing.T, api *fakeTaskAPI) (*draftmcp.HTTPClient, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(api.handler(t))
	t.Cleanup(srv.Close)
	return draftmcp.New(srv.URL, api.sessionID, api.nonce, srv.Client()), srv
}

func TestHTTPClient_GetSession(t *testing.T) {
	api := newFakeTaskAPI()
	c, _ := newFakeServer(t, api)
	view, err := c.GetSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if view.WorktreeID != api.worktreeID {
		t.Fatalf("worktree=%q", view.WorktreeID)
	}
	if view.Snapshot.Prompt != api.prompt {
		t.Fatalf("prompt=%q", view.Snapshot.Prompt)
	}
}

func TestHTTPClient_SetPrompt_nonceRequired(t *testing.T) {
	api := newFakeTaskAPI()
	srv := httptest.NewServer(api.handler(t))
	defer srv.Close()
	badClient := draftmcp.New(srv.URL, api.sessionID, "", srv.Client())
	err := badClient.SetPrompt(context.Background(), "<p>x</p>")
	if err == nil {
		t.Fatal("expected error when nonce is empty")
	}
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("err=%v; want ErrUnauthorized", err)
	}
}

func TestHTTPClient_SetPrompt_staleNonce(t *testing.T) {
	api := newFakeTaskAPI()
	srv := httptest.NewServer(api.handler(t))
	defer srv.Close()
	bad := draftmcp.New(srv.URL, api.sessionID, "wrong", srv.Client())
	err := bad.SetPrompt(context.Background(), "<p>ok</p>")
	if err == nil || !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized; got %v", err)
	}
	if api.patchCalls != 0 {
		t.Fatalf("stale patch reached server: calls=%d", api.patchCalls)
	}
}

func TestHTTPClient_SetPrompt_ok(t *testing.T) {
	api := newFakeTaskAPI()
	c, _ := newFakeServer(t, api)
	if err := c.SetPrompt(context.Background(), "<h2>New</h2>"); err != nil {
		t.Fatal(err)
	}
	if api.prompt != "<h2>New</h2>" {
		t.Fatalf("server prompt=%q", api.prompt)
	}
	if api.patchCalls != 1 {
		t.Fatalf("patchCalls=%d", api.patchCalls)
	}
}

func TestHTTPClient_SetPrompt_serverValidatorRejectsScript(t *testing.T) {
	api := newFakeTaskAPI()
	c, _ := newFakeServer(t, api)
	err := c.SetPrompt(context.Background(), "<p><script>alert(1)</script></p>")
	if err == nil {
		t.Fatal("expected server-side reject")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err=%v; want ErrInvalidInput", err)
	}
	if api.patchCalls != 0 {
		t.Fatalf("bad payload reached server: patchCalls=%d", api.patchCalls)
	}
	if api.prompt != "<p>hello</p>" {
		t.Fatalf("server prompt changed: %q", api.prompt)
	}
}

func TestHTTPClient_SearchRepoFiles(t *testing.T) {
	api := newFakeTaskAPI()
	c, _ := newFakeServer(t, api)
	page, err := c.SearchRepoFiles(context.Background(), draftmcp.SearchRepoFilesInput{
		WorktreeID: api.worktreeID,
		Query:      "draft",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Paths) != 2 {
		t.Fatalf("paths=%v", page.Paths)
	}
}

func TestHTTPClient_ReadRepoFile(t *testing.T) {
	api := newFakeTaskAPI()
	c, _ := newFakeServer(t, api)
	fp, err := c.ReadRepoFile(context.Background(), api.worktreeID, "README.md")
	if err != nil {
		t.Fatal(err)
	}
	if fp.Content == "" || fp.LineCount == 0 {
		t.Fatalf("bad file: %+v", fp)
	}
}

func TestHTTPClient_ListTemplates(t *testing.T) {
	api := newFakeTaskAPI()
	c, _ := newFakeServer(t, api)
	raw, err := c.ListTemplates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["templates"]; !ok {
		t.Fatalf("no templates key: %v", raw)
	}
}

func TestHTTPClient_SearchTasks_filter(t *testing.T) {
	api := newFakeTaskAPI()
	c, _ := newFakeServer(t, api)
	hits, err := c.SearchTasks(context.Background(), "login", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].ID != "t-2" {
		t.Fatalf("hits=%+v", hits)
	}
}

func TestHTTPClient_SearchTasks_limit(t *testing.T) {
	api := newFakeTaskAPI()
	c, _ := newFakeServer(t, api)
	hits, err := c.SearchTasks(context.Background(), "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("limit ignored: %d", len(hits))
	}
}

func TestWriteBind_roundtrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bind.json")
	err := draftmcp.WriteBind(p, draftmcp.BindFile{
		SessionID:      "s",
		Nonce:          "n",
		TaskAPIBaseURL: "http://taskapi",
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var got draftmcp.BindFile
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.BindSchemaVersion != draftmcp.BindSchemaVersion {
		t.Fatalf("schema=%d", got.BindSchemaVersion)
	}
	if got.SessionID != "s" || got.Nonce != "n" || got.TaskAPIBaseURL != "http://taskapi" {
		t.Fatalf("roundtrip lost fields: %+v", got)
	}
	loaded, err := draftmcp.LoadBind(p)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SessionID != "s" {
		t.Fatalf("loaded=%+v", loaded)
	}
}

func TestWriteBind_missingRequired(t *testing.T) {
	dir := t.TempDir()
	err := draftmcp.WriteBind(filepath.Join(dir, "b.json"), draftmcp.BindFile{})
	if err == nil {
		t.Fatal("expected error when session/nonce missing")
	}
}

// TestToolHost_writePromptRejectsScript ensures the tool-side validator
// blocks disallowed HTML before it ever hits the HTTP client. This is the
// belt-and-suspenders leg that guarantees ErrInvalidInput bubbles up even
// when a compromised server would accept it.
func TestToolHost_writePromptRejectsScript(t *testing.T) {
	api := newFakeTaskAPI()
	c, _ := newFakeServer(t, api)
	host := &draftmcp.ToolHost{
		Bind:   &draftmcp.BindFile{SessionID: api.sessionID, Nonce: api.nonce, TaskAPIBaseURL: c.BaseURL()},
		Client: c,
	}
	err := host.WritePromptForTest(context.Background(), "<script>alert(1)</script>")
	if err == nil {
		t.Fatal("expected validator to reject")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err=%v; want ErrInvalidInput", err)
	}
	if api.patchCalls != 0 {
		t.Fatalf("bad payload leaked to server: calls=%d", api.patchCalls)
	}
}
