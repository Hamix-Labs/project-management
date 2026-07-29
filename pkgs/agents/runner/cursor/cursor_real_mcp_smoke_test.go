//go:build cursor_real

package cursor_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/agentmcp"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/cursor"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// TestCursorAdapter_RealBinary_MCPSubmitCriteria is the Plan 3 merge-gate smoke:
// Cursor headless with --approve-mcps --trust loads workspace .cursor/mcp.json
// and can call hamix.submit_criteria_report.
func TestCursorAdapter_RealBinary_MCPSubmitCriteria(t *testing.T) {
	if os.Getenv(realCursorRunGateEnv) != "1" {
		t.Skipf("skipping: %s != 1", realCursorRunGateEnv)
	}
	mcpBin, err := exec.LookPath("hamix-agent-mcp")
	if err != nil {
		t.Fatalf("hamix-agent-mcp not on PATH (build cmd/hamix-agent-mcp first): %v", err)
	}
	binaryPath := os.Getenv(realCursorBinaryEnv)
	if binaryPath == "" {
		binaryPath = "cursor-agent"
	}
	probeCtx, probeCancel := context.WithTimeout(context.Background(), cursor.DefaultProbeTimeout)
	defer probeCancel()
	version, probeErr := cursor.Probe(probeCtx, binaryPath, cursor.DefaultProbeTimeout, nil)
	if probeErr != nil {
		t.Fatalf("cursor probe failed: %v", probeErr)
	}

	workDir := t.TempDir()
	reportDir := t.TempDir()
	cycleID := "mcp-smoke-1"
	if err := sidecar.EnsureReportCycleDir(reportDir, cycleID); err != nil {
		t.Fatal(err)
	}
	bindPath := agentmcp.BindPath(reportDir, cycleID)
	if err := agentmcp.WriteBind(bindPath, agentmcp.BindFile{
		TaskID:             "task-smoke",
		CycleID:            cycleID,
		Phase:              agentmcp.PhaseExecute,
		ReportDir:          reportDir,
		WorkingDir:         workDir,
		ActiveCriterionIDs: []string{"c1"},
		SubmitNonce:        "smoke-nonce-1",
	}); err != nil {
		t.Fatal(err)
	}
	mcpDir := filepath.Join(workDir, ".cursor")
	if err := os.MkdirAll(mcpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := map[string]any{
		"mcpServers": map[string]any{
			"hamix-agent": map[string]any{
				"type":    "stdio",
				"command": mcpBin,
				"args":    []string{"--bind", bindPath},
			},
		},
	}
	raw, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(filepath.Join(mcpDir, "mcp.json"), append(raw, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := cursor.New(cursor.Options{BinaryPath: binaryPath, Version: version})
	runCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	prompt := strings.TrimSpace(`
Call the MCP tool hamix.submit_criteria_report exactly once with:
{"criteria":[{"id":"c1","claimed_done":true,"evidence":"mcp smoke evidence for c1"}]}
Do not write any files yourself. After the tool succeeds, reply with the single word DONE.
`)
	res, err := adapter.Run(runCtx, runner.Request{
		TaskID:         "task-smoke",
		AttemptSeq:     1,
		Phase:          cyclesdomain.PhaseExecute,
		Prompt:         prompt,
		WorkingDir:     workDir,
		Timeout:        3 * time.Minute,
		ApproveMCPs:    true,
		TrustWorkspace: true,
	})
	if err != nil {
		t.Fatalf("Run: %v (summary=%q)", err, res.Summary)
	}
	if err := sidecar.RequireCriteriaSubmitReceipt(reportDir, cycleID, "smoke-nonce-1"); err != nil {
		t.Fatalf("receipt missing after MCP submit: %v; summary=%q raw=%s", err, res.Summary, res.RawOutput)
	}
	if _, err := sidecar.ParseCriteriaReport(reportDir, cycleID, map[string]struct{}{"c1": {}}); err != nil {
		t.Fatalf("criteria report invalid: %v", err)
	}
}
