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

// ToolHost is the in-process store surface MCP tools use in tests and
// when the binary is launched with an injected store (dev).
type ToolHost struct {
	Bind  *BindFile
	Store contract.Store
}

// DefaultTools registers prompt-write and read tools on the MCP server.
func RegisterTools(server *mcp.Server, host *ToolHost) {
	type emptyIn struct{}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hamix.draft_get",
		Description: "Read the current compose form snapshot for this draft-assist session.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, domain.FormSnapshot, error) {
		sess, err := host.Store.GetSession(ctx, host.Bind.SessionID)
		if err != nil {
			return nil, domain.FormSnapshot{}, err
		}
		return nil, sess.Snapshot, nil
	})

	type setPromptIn struct {
		Prompt string `json:"prompt"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hamix.draft_set_prompt",
		Description: "Replace the initial prompt HTML for this session.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in setPromptIn) (*mcp.CallToolResult, map[string]string, error) {
		if _, err := host.Store.UpdatePrompt(ctx, host.Bind.SessionID, host.Bind.Nonce, in.Prompt); err != nil {
			return nil, nil, err
		}
		_ = host.Store.Publish(ctx, host.Bind.SessionID, domain.Event{
			Kind: domain.EventPatch,
			Data: domain.PatchEventData{Op: domain.PatchOpSet, Value: in.Prompt, Summary: "Prompt updated"},
		})
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
		sess, err := host.Store.GetSession(ctx, host.Bind.SessionID)
		if err != nil {
			return nil, nil, err
		}
		prompt := sess.Snapshot.Prompt
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
		if _, err := host.Store.UpdatePrompt(ctx, host.Bind.SessionID, host.Bind.Nonce, prompt); err != nil {
			return nil, nil, err
		}
		_ = host.Store.Publish(ctx, host.Bind.SessionID, domain.Event{
			Kind: domain.EventPatch,
			Data: domain.PatchEventData{Op: op, Find: in.Find, Value: in.Value, Summary: "Prompt patched"},
		})
		return nil, map[string]string{"status": "ok"}, nil
	})
}

// RunStdio serves MCP over stdin/stdout with DefaultTools.
func RunStdio(ctx context.Context, host *ToolHost) error {
	server := mcp.NewServer(&mcp.Implementation{Name: "hamix-draft-mcp", Version: "v1.0.0"}, nil)
	RegisterTools(server, host)
	return server.Run(ctx, &mcp.StdioTransport{})
}
