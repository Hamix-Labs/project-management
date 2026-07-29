package harness

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/agentmcp"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

const (
	agentMCPBinaryName = "hamix-agent-mcp"
	agentMCPServerName = "hamix-agent"
)

// agentMCPActive reports whether this run should use tool-only MCP submit.
// Non-cursor runners (fake tests) never activate MCP. Production Cursor path
// follows app_settings.agent_mcp_enabled (default true) unless Options overrides.
func (h *Harness) agentMCPActive(ctx context.Context) bool {
	if h == nil {
		return false
	}
	if h.opts.AgentMCPEnabled != nil {
		return *h.opts.AgentMCPEnabled
	}
	if !isCursorSessionRunner(h.runner) {
		return false
	}
	settings, err := h.store.GetSettings(ctx)
	if err != nil {
		// Fail closed to product default (tool-only) when settings are unreadable
		// on a Cursor runner — never silently fall back to freeform.
		return true
	}
	return settings.AgentMCPEnabled
}

type agentMCPPrep struct {
	Nonce    string
	BindPath string
}

func mintAgentMCPNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func resolveAgentMCPBinary() (string, error) {
	return exec.LookPath(agentMCPBinaryName)
}

func sortedIDList(ids map[string]struct{}) []string {
	out := make([]string, 0, len(ids))
	for id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func lockedCriterionIDSet(state *processState) map[string]struct{} {
	if state == nil || len(state.verify.previouslyPassed) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(state.verify.previouslyPassed))
	for id := range state.verify.previouslyPassed {
		out[id] = struct{}{}
	}
	return out
}

func workspaceMCPConfigPath(workingDir string) string {
	return filepath.Join(strings.TrimSpace(workingDir), ".cursor", "mcp.json")
}

// prepareAgentMCP writes bind under ReportDir and mcp.json under the Cursor
// workspace (WorkingDir). Cursor CLI only loads project MCP from
// <workspace>/.cursor/mcp.json — --add-dir does not discover MCP config.
func (h *Harness) prepareAgentMCP(
	ctx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	phase cyclesdomain.Phase,
	state *processState,
) (agentMCPPrep, error) {
	_ = ctx
	bin, err := resolveAgentMCPBinary()
	if err != nil {
		return agentMCPPrep{}, fmt.Errorf("agent MCP binary %q not found on PATH: %w", agentMCPBinaryName, err)
	}
	nonce, err := mintAgentMCPNonce()
	if err != nil {
		return agentMCPPrep{}, fmt.Errorf("mint submit nonce: %w", err)
	}
	if err := reports.EnsureReportCycleDir(h.opts.ReportDir, cycle.ID); err != nil {
		return agentMCPPrep{}, err
	}
	phaseName := agentmcp.PhaseExecute
	if phase == cyclesdomain.PhaseVerify {
		phaseName = agentmcp.PhaseVerify
	}
	active := expectedActiveCriterionIDs(state)
	bind := agentmcp.BindFile{
		TaskID:             task.ID,
		CycleID:            cycle.ID,
		Phase:              phaseName,
		ReportDir:          h.opts.ReportDir,
		WorkingDir:         h.opts.WorkingDir,
		ActiveCriterionIDs: sortedIDList(active),
		LockedCriterionIDs: sortedIDList(lockedCriterionIDSet(state)),
		SubmitNonce:        nonce,
	}
	bindPath := agentmcp.BindPath(h.opts.ReportDir, cycle.ID)
	if err := agentmcp.WriteBind(bindPath, bind); err != nil {
		return agentMCPPrep{}, fmt.Errorf("write agent MCP bind: %w", err)
	}
	if err := h.writeWorkspaceMCPConfig(bin, bindPath, state); err != nil {
		return agentMCPPrep{}, err
	}
	return agentMCPPrep{Nonce: nonce, BindPath: bindPath}, nil
}

func (h *Harness) writeWorkspaceMCPConfig(bin, bindPath string, state *processState) error {
	wd := strings.TrimSpace(h.opts.WorkingDir)
	if wd == "" {
		return fmt.Errorf("working dir is empty; cannot write workspace MCP config")
	}
	mcpPath := workspaceMCPConfigPath(wd)
	if err := os.MkdirAll(filepath.Dir(mcpPath), 0o755); err != nil {
		return err
	}
	if state != nil && !state.agentMCP.mcpConfigTracked {
		prev, err := os.ReadFile(mcpPath)
		if err == nil {
			state.agentMCP.mcpConfigBackup = append([]byte(nil), prev...)
			state.agentMCP.mcpConfigHadFile = true
		} else if !os.IsNotExist(err) {
			return err
		}
		state.agentMCP.mcpConfigTracked = true
	}

	servers := map[string]any{}
	if state != nil && state.agentMCP.mcpConfigHadFile {
		var existing map[string]any
		if err := json.Unmarshal(state.agentMCP.mcpConfigBackup, &existing); err == nil {
			if raw, ok := existing["mcpServers"].(map[string]any); ok {
				servers = raw
			}
		}
	}
	servers[agentMCPServerName] = map[string]any{
		"type":    "stdio",
		"command": bin,
		"args":    []string{"--bind", bindPath},
	}
	cfg := map[string]any{"mcpServers": servers}
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(mcpPath, raw, 0o644); err != nil {
		return fmt.Errorf("write workspace MCP config: %w", err)
	}
	return nil
}

// restoreWorkspaceMCPConfig reverts WorkingDir/.cursor/mcp.json after the cycle.
func (h *Harness) restoreWorkspaceMCPConfig(state *processState) {
	if h == nil || state == nil || !state.agentMCP.mcpConfigTracked {
		return
	}
	mcpPath := workspaceMCPConfigPath(h.opts.WorkingDir)
	if mcpPath == ".cursor"+string(filepath.Separator)+"mcp.json" || strings.TrimSpace(h.opts.WorkingDir) == "" {
		return
	}
	if state.agentMCP.mcpConfigHadFile {
		_ = os.WriteFile(mcpPath, state.agentMCP.mcpConfigBackup, 0o644)
		return
	}
	_ = os.Remove(mcpPath)
	// Best-effort: remove empty .cursor dir we may have created.
	_ = os.Remove(filepath.Dir(mcpPath))
}

func applyAgentMCPToRequest(req *runner.Request, _ agentMCPPrep) {
	req.ApproveMCPs = true
	req.TrustWorkspace = true
}

// mcpPrepareFailedResult is returned when bind/MCP launch config cannot be written.
func mcpPrepareFailedResult(err error) (runner.Result, error) {
	return runner.NewResult(cyclesdomain.PhaseStatusFailed, "agent MCP prepare failed: "+err.Error(), nil, ""), err
}
