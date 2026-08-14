package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
)

// NonceHeader is the request header carrying the session-scoped MCP nonce.
// See docs/domain/draft-assist.md and ADR-0101.
const NonceHeader = "X-Hamix-Draft-Nonce"

// HTTPClient talks to a running taskapi as one draft-assist MCP host. It
// carries a session-scoped nonce on every write and translates HTTP status
// codes into the domain sentinels used by tool handlers.
//
// A zero HTTPClient is not usable; construct with New.
type HTTPClient struct {
	baseURL   string
	nonce     string
	sessionID string
	hc        *http.Client
}

// SessionView is the subset of session state MCP tools need. Mirrors the
// taskapi GET /draft-assist/sessions/{id} response.
type SessionView struct {
	ID         string              `json:"id"`
	Nonce      string              `json:"nonce"`
	WorktreeID string              `json:"worktree_id"`
	Snapshot   domain.FormSnapshot `json:"snapshot"`
}

// New returns a client bound to a single session (id + nonce). baseURL is
// the taskapi origin (no trailing slash required). If httpClient is nil the
// default client with a 30s timeout is used.
//
//funclogmeasure:skip category=hot-path reason="Constructor; call sites emit the operation trace."
func New(baseURL, sessionID, nonce string, httpClient *http.Client) *HTTPClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &HTTPClient{
		baseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		sessionID: strings.TrimSpace(sessionID),
		nonce:     strings.TrimSpace(nonce),
		hc:        httpClient,
	}
}

// SessionID returns the bound session id.
//
//funclogmeasure:skip category=hot-path reason="Pure accessor."
func (c *HTTPClient) SessionID() string { return c.sessionID }

// Nonce returns the bound nonce (used by tests and diagnostics).
//
//funclogmeasure:skip category=hot-path reason="Pure accessor."
func (c *HTTPClient) Nonce() string { return c.nonce }

// BaseURL returns the taskapi origin the client talks to.
//
//funclogmeasure:skip category=hot-path reason="Pure accessor."
func (c *HTTPClient) BaseURL() string { return c.baseURL }

// GetSession returns the current session view for the bound session id.
func (c *HTTPClient) GetSession(ctx context.Context) (*SessionView, error) {
	var out SessionView
	if err := c.do(ctx, http.MethodGet, "/draft-assist/sessions/"+c.sessionID, nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetPrompt replaces the prompt HTML on the bound session. The nonce header
// is required by the server; a stale nonce yields ErrUnauthorized.
func (c *HTTPClient) SetPrompt(ctx context.Context, prompt string) error {
	body := map[string]string{"prompt": prompt}
	return c.do(ctx, http.MethodPatch, "/draft-assist/sessions/"+c.sessionID+"/prompt", body, true, nil)
}

// SearchRepoFilesInput scopes a repo file search to one worktree.
type SearchRepoFilesInput struct {
	WorktreeID string
	Query      string
	Limit      int
	After      string
}

// RepoFilesPage mirrors the taskapi GET /repo/files response shape.
type RepoFilesPage struct {
	Paths     []string `json:"paths"`
	HasMore   bool     `json:"has_more"`
	Truncated bool     `json:"truncated"`
	Source    string   `json:"source,omitempty"`
	NextAfter string   `json:"next_after,omitempty"`
}

// SearchRepoFiles calls GET /repo/files scoped to worktree_id.
func (c *HTTPClient) SearchRepoFiles(ctx context.Context, in SearchRepoFilesInput) (*RepoFilesPage, error) {
	if strings.TrimSpace(in.WorktreeID) == "" {
		return nil, fmt.Errorf("%w: worktree_id required for repo search", domain.ErrInvalidInput)
	}
	q := url.Values{}
	q.Set("worktree_id", in.WorktreeID)
	if strings.TrimSpace(in.Query) != "" {
		q.Set("q", in.Query)
	}
	if in.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", in.Limit))
	}
	if strings.TrimSpace(in.After) != "" {
		q.Set("after", in.After)
	}
	var out RepoFilesPage
	if err := c.do(ctx, http.MethodGet, "/repo/files?"+q.Encode(), nil, false, &out); err != nil {
		return nil, err
	}
	if out.Paths == nil {
		out.Paths = []string{}
	}
	return &out, nil
}

// RepoFile mirrors the taskapi GET /repo/file response.
type RepoFile struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Binary    bool   `json:"binary"`
	Truncated bool   `json:"truncated"`
	SizeBytes int64  `json:"size_bytes"`
	LineCount int    `json:"line_count"`
	Warning   string `json:"warning,omitempty"`
}

// ReadRepoFile calls GET /repo/file for the given worktree/path.
func (c *HTTPClient) ReadRepoFile(ctx context.Context, worktreeID, path string) (*RepoFile, error) {
	if strings.TrimSpace(worktreeID) == "" {
		return nil, fmt.Errorf("%w: worktree_id required for repo read", domain.ErrInvalidInput)
	}
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("%w: path required", domain.ErrInvalidInput)
	}
	q := url.Values{}
	q.Set("worktree_id", worktreeID)
	q.Set("path", path)
	var out RepoFile
	if err := c.do(ctx, http.MethodGet, "/repo/file?"+q.Encode(), nil, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TemplateSummary is the trimmed template row surfaced to MCP callers.
type TemplateSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListTemplates calls GET /task-templates and returns the raw JSON. Servers
// return `{ templates: [...] }` today; we forward that shape so callers can
// evolve independently of taskapi response tweaks.
func (c *HTTPClient) ListTemplates(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodGet, "/task-templates", nil, false, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

// TaskSummary is a trimmed task row surfaced to MCP callers.
type TaskSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// SearchTasks calls GET /tasks and filters locally by title matching q. The
// taskapi list endpoint does not accept a `q=` query yet; local filtering
// keeps the tool honest while giving the LLM the same client-side match
// semantics as the SPA task list.
func (c *HTTPClient) SearchTasks(ctx context.Context, q string, limit int) ([]TaskSummary, error) {
	var raw map[string]any
	values := url.Values{}
	if limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", limit))
	}
	path := "/tasks"
	if e := values.Encode(); e != "" {
		path += "?" + e
	}
	if err := c.do(ctx, http.MethodGet, path, nil, false, &raw); err != nil {
		return nil, err
	}
	needle := strings.ToLower(strings.TrimSpace(q))
	items, _ := raw["tasks"].([]any)
	out := make([]TaskSummary, 0, len(items))
	for _, it := range items {
		row, ok := it.(map[string]any)
		if !ok {
			continue
		}
		id, _ := row["id"].(string)
		title, _ := row["title"].(string)
		if needle != "" && !strings.Contains(strings.ToLower(title), needle) {
			continue
		}
		out = append(out, TaskSummary{ID: id, Title: title})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// do performs one request and maps taskapi status codes onto domain sentinels.
// When withNonce is true the request carries the X-Hamix-Draft-Nonce header
// bound to this client. dst may be nil (response ignored) or a pointer to a
// value the response is JSON-decoded into.
func (c *HTTPClient) do(ctx context.Context, method, path string, body any, withNonce bool, dst any) error {
	if c.baseURL == "" {
		return fmt.Errorf("%w: taskapi base URL not bound", domain.ErrUnavailable)
	}
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if withNonce {
		if c.nonce == "" {
			return fmt.Errorf("%w: MCP nonce not bound", domain.ErrUnauthorized)
		}
		req.Header.Set(NonceHeader, c.nonce)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return c.errorFromStatus(resp)
	}
	if dst == nil {
		return nil
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(dst); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure status mapping; do() emits the operation trace."
func (c *HTTPClient) errorFromStatus(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	msg := strings.TrimSpace(string(body))
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("%w: taskapi %d %s", domain.ErrUnauthorized, resp.StatusCode, msg)
	case http.StatusNotFound:
		return fmt.Errorf("%w: taskapi 404 %s", domain.ErrNotFound, msg)
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return fmt.Errorf("%w: taskapi %d %s", domain.ErrInvalidInput, resp.StatusCode, msg)
	case http.StatusConflict:
		return fmt.Errorf("%w: taskapi 409 %s", domain.ErrRunActive, msg)
	default:
		return fmt.Errorf("taskapi %d: %s", resp.StatusCode, msg)
	}
}
