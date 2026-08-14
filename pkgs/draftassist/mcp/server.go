package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const BindSchemaVersion = 1

// BindFile is the on-disk contract for hamix-draft-mcp.
type BindFile struct {
	BindSchemaVersion int    `json:"bind_schema_version"`
	SessionID         string `json:"session_id"`
	Nonce             string `json:"nonce"`
	TaskAPIBaseURL    string `json:"taskapi_base_url,omitempty"`
}

// LoadBind reads and validates a draft-assist bind file.
//
//funclogmeasure:skip category=hot-path reason="Bind-file parse helper; host emits operation traces."
func LoadBind(path string) (*BindFile, error) {
	raw, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return nil, fmt.Errorf("read bind: %w", err)
	}
	var b BindFile
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("parse bind: %w", err)
	}
	if b.BindSchemaVersion != BindSchemaVersion {
		return nil, fmt.Errorf("unsupported bind_schema_version %d", b.BindSchemaVersion)
	}
	if strings.TrimSpace(b.SessionID) == "" || strings.TrimSpace(b.Nonce) == "" {
		return nil, fmt.Errorf("session_id and nonce are required")
	}
	return &b, nil
}

// PromptClient is the subset of HTTPClient the MCP tool bodies rely on.
// It is satisfied by *HTTPClient in production and by a fake in tests.
type PromptClient interface {
	GetSession(ctx context.Context) (*SessionView, error)
	SetPrompt(ctx context.Context, prompt string) error
	SearchRepoFiles(ctx context.Context, in SearchRepoFilesInput) (*RepoFilesPage, error)
	ReadRepoFile(ctx context.Context, worktreeID, path string) (*RepoFile, error)
	ListTemplates(ctx context.Context) (map[string]any, error)
	SearchTasks(ctx context.Context, q string, limit int) ([]TaskSummary, error)
}

// ToolHost is the surface MCP tool handlers use to reach the outside world.
//
// One of Client or Store must be set. In production (hamix-draft-mcp bound
// via --bind) Client talks to taskapi HTTP. In tests and in-process smoke
// paths Store is used directly. When both are set Client wins for writes,
// so tests can dependency-inject a fake.
type ToolHost struct {
	Bind   *BindFile
	Client PromptClient
	Store  contract.Store
}

// snapshotView normalises a session lookup across the two host modes.
func (h *ToolHost) snapshotView(ctx context.Context) (*SessionView, error) {
	if h.Client != nil {
		return h.Client.GetSession(ctx)
	}
	sess, err := h.Store.GetSession(ctx, h.Bind.SessionID)
	if err != nil {
		return nil, err
	}
	return &SessionView{
		ID:         sess.ID,
		Nonce:      sess.Nonce,
		WorktreeID: sess.WorktreeID,
		Snapshot:   sess.Snapshot,
	}, nil
}

// writePrompt runs the shared write path: validate → dispatch (HTTP or store)
// → publish patch event when using the in-process store.
func (h *ToolHost) writePrompt(ctx context.Context, prompt string, ev domain.PatchEventData) error {
	if err := domain.ValidateHTML(prompt); err != nil {
		return err
	}
	if h.Client != nil {
		return h.Client.SetPrompt(ctx, prompt)
	}
	if _, err := h.Store.UpdatePrompt(ctx, h.Bind.SessionID, h.Bind.Nonce, prompt); err != nil {
		return err
	}
	_ = h.Store.Publish(ctx, h.Bind.SessionID, domain.Event{
		Kind: domain.EventPatch,
		Data: ev,
	})
	return nil
}

// RegisterTools wires the v1 draft-assist tool table onto server.
//
//funclogmeasure:skip category=hot-path reason="Tool table wiring; tool handlers emit traces when executed."
func RegisterTools(server *mcp.Server, host *ToolHost) {
	type emptyIn struct{}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hamix.draft_get",
		Description: "Read the current compose form snapshot for this draft-assist session.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, domain.FormSnapshot, error) {
		view, err := host.snapshotView(ctx)
		if err != nil {
			return nil, domain.FormSnapshot{}, err
		}
		return nil, view.Snapshot, nil
	})

	type setPromptIn struct {
		Prompt string `json:"prompt"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hamix.draft_set_prompt",
		Description: "Replace the initial prompt HTML for this session (validated TipTap subset).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in setPromptIn) (*mcp.CallToolResult, map[string]string, error) {
		if err := host.writePrompt(ctx, in.Prompt, domain.PatchEventData{
			Op:      domain.PatchOpSet,
			Value:   in.Prompt,
			Summary: "Prompt updated",
		}); err != nil {
			return nil, nil, err
		}
		return nil, map[string]string{"status": "ok"}, nil
	})

	type patchPromptIn struct {
		Op    string `json:"op"`
		Find  string `json:"find,omitempty"`
		Value string `json:"value"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hamix.draft_patch_prompt",
		Description: "Bounded find/replace or append on the prompt. Prefer over full replace.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in patchPromptIn) (*mcp.CallToolResult, map[string]string, error) {
		view, err := host.snapshotView(ctx)
		if err != nil {
			return nil, nil, err
		}
		prompt := view.Snapshot.Prompt
		op := domain.PatchOp(in.Op)
		switch op {
		case domain.PatchOpAppend:
			prompt = prompt + in.Value
		case domain.PatchOpFindReplace:
			if in.Find == "" {
				return nil, nil, fmt.Errorf("%w: find required", domain.ErrInvalidInput)
			}
			prompt = strings.Replace(prompt, in.Find, in.Value, 1)
		case domain.PatchOpSet:
			prompt = in.Value
		default:
			return nil, nil, fmt.Errorf("%w: unknown op %q", domain.ErrInvalidInput, in.Op)
		}
		if err := host.writePrompt(ctx, prompt, domain.PatchEventData{
			Op:      op,
			Find:    in.Find,
			Value:   in.Value,
			Summary: "Prompt patched",
		}); err != nil {
			return nil, nil, err
		}
		return nil, map[string]string{"status": "ok"}, nil
	})

	type searchRepoIn struct {
		Query string `json:"query"`
		Limit int    `json:"limit,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hamix.draft_search_repo",
		Description: "Search the bound worktree for paths matching a query (read-only).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in searchRepoIn) (*mcp.CallToolResult, map[string]any, error) {
		if host.Client == nil {
			return nil, nil, fmt.Errorf("%w: taskapi client not bound", domain.ErrUnavailable)
		}
		view, err := host.snapshotView(ctx)
		if err != nil {
			return nil, nil, err
		}
		page, err := host.Client.SearchRepoFiles(ctx, SearchRepoFilesInput{
			WorktreeID: view.WorktreeID,
			Query:      in.Query,
			Limit:      in.Limit,
		})
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{
			"paths":      page.Paths,
			"has_more":   page.HasMore,
			"truncated":  page.Truncated,
			"next_after": page.NextAfter,
		}, nil
	})

	type readFileIn struct {
		Path string `json:"path"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hamix.draft_read_file",
		Description: "Read a file from the bound worktree (read-only).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in readFileIn) (*mcp.CallToolResult, map[string]any, error) {
		if strings.TrimSpace(in.Path) == "" {
			return nil, nil, fmt.Errorf("%w: path required", domain.ErrInvalidInput)
		}
		if host.Client == nil {
			return nil, nil, fmt.Errorf("%w: taskapi client not bound", domain.ErrUnavailable)
		}
		view, err := host.snapshotView(ctx)
		if err != nil {
			return nil, nil, err
		}
		fp, err := host.Client.ReadRepoFile(ctx, view.WorktreeID, in.Path)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{
			"path":       fp.Path,
			"content":    fp.Content,
			"binary":     fp.Binary,
			"truncated":  fp.Truncated,
			"size_bytes": fp.SizeBytes,
			"line_count": fp.LineCount,
			"warning":    fp.Warning,
		}, nil
	})

	type listTemplatesIn struct {
		Query string `json:"query,omitempty"`
		Limit int    `json:"limit,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hamix.draft_list_templates",
		Description: "List saved task templates the operator can reuse.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listTemplatesIn) (*mcp.CallToolResult, map[string]any, error) {
		if host.Client == nil {
			return nil, nil, fmt.Errorf("%w: taskapi client not bound", domain.ErrUnavailable)
		}
		raw, err := host.Client.ListTemplates(ctx)
		if err != nil {
			return nil, nil, err
		}
		return nil, raw, nil
	})

	type searchTasksIn struct {
		Query string `json:"query"`
		Limit int    `json:"limit,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hamix.draft_search_tasks",
		Description: "Search existing tasks for context while drafting (title-substring, read-only).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in searchTasksIn) (*mcp.CallToolResult, map[string]any, error) {
		if host.Client == nil {
			return nil, nil, fmt.Errorf("%w: taskapi client not bound", domain.ErrUnavailable)
		}
		hits, err := host.Client.SearchTasks(ctx, in.Query, in.Limit)
		if err != nil {
			return nil, nil, err
		}
		out := make([]map[string]any, 0, len(hits))
		for _, t := range hits {
			out = append(out, map[string]any{"id": t.ID, "title": t.Title})
		}
		return nil, map[string]any{"tasks": out}, nil
	})
}

// RunStdio serves MCP over stdin/stdout with DefaultTools.
//
//funclogmeasure:skip category=hot-path reason="Stdio bootstrap; MCP SDK owns request traces."
func RunStdio(ctx context.Context, host *ToolHost) error {
	server := mcp.NewServer(&mcp.Implementation{Name: "hamix-draft-mcp", Version: "v1.0.0"}, nil)
	RegisterTools(server, host)
	return server.Run(ctx, &mcp.StdioTransport{})
}
