package harness

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/agentmcp"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/execute"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestProbeCriteriaReport_requiresReceiptWhenEnabled(t *testing.T) {
	t.Parallel()
	reportDir := t.TempDir()
	cycleID := "cycle-mcp-1"
	expected := map[string]struct{}{"c1": {}}
	if err := sidecar.WriteCriteriaReport(reportDir, cycleID, []sidecar.CriteriaEntry{
		{ID: "c1", ClaimedDone: true, Evidence: "done"},
	}, nil); err != nil {
		t.Fatal(err)
	}
	if msg := execute.ProbeCriteriaReportWithReceipt(reportDir, cycleID, expected, true, "nonce-a"); msg == "" {
		t.Fatal("expected probe failure without receipt")
	}
	if err := sidecar.WriteSubmitReceipt(sidecar.CriteriaSubmitReceiptPath(reportDir, cycleID), sidecar.SubmitReceipt{
		Nonce:   "nonce-a",
		Phase:   "execute",
		CycleID: cycleID,
		Tool:    "hamix.submit_criteria_report",
	}); err != nil {
		t.Fatal(err)
	}
	if msg := execute.ProbeCriteriaReportWithReceipt(reportDir, cycleID, expected, true, "nonce-a"); msg != "" {
		t.Fatalf("expected probe ok with matching receipt; got %q", msg)
	}
	if msg := execute.ProbeCriteriaReportWithReceipt(reportDir, cycleID, expected, true, "nonce-b"); msg == "" {
		t.Fatal("expected probe failure on nonce mismatch")
	}
	if msg := execute.ProbeCriteriaReport(reportDir, cycleID, expected); msg != "" {
		t.Fatalf("legacy probe should accept report without receipt; got %q", msg)
	}
}

func TestPrepareAgentMCP_writesBindAndWorkspaceMCPConfig(t *testing.T) {
	binDir := t.TempDir()
	name := agentMCPBinaryName
	if runtime.GOOS == "windows" {
		name += ".cmd"
		if err := os.WriteFile(filepath.Join(binDir, name), []byte("@echo off\r\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	reportDir := t.TempDir()
	workDir := t.TempDir()
	h := &Harness{opts: Options{ReportDir: reportDir, WorkingDir: workDir}}
	task := &taskcoredomain.Task{ID: "task-1"}
	cycle := &cyclesdomain.TaskCycle{ID: "cycle-1"}
	state := &processState{}
	prep, err := h.prepareAgentMCP(context.Background(), task, cycle, cyclesdomain.PhaseExecute, state)
	if err != nil {
		t.Fatalf("prepareAgentMCP: %v", err)
	}
	if prep.Nonce == "" {
		t.Fatal("expected nonce")
	}
	if _, err := os.Stat(agentmcp.BindPath(reportDir, cycle.ID)); err != nil {
		t.Fatalf("bind missing: %v", err)
	}
	mcpPath := workspaceMCPConfigPath(workDir)
	raw, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("workspace mcp.json missing: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	servers, _ := cfg["mcpServers"].(map[string]any)
	if _, ok := servers[agentMCPServerName]; !ok {
		t.Fatalf("missing %s in mcp.json: %s", agentMCPServerName, raw)
	}
	req := runner.Request{}
	applyAgentMCPToRequest(&req, prep)
	if !req.ApproveMCPs || !req.TrustWorkspace {
		t.Fatalf("flags: ApproveMCPs=%v TrustWorkspace=%v", req.ApproveMCPs, req.TrustWorkspace)
	}
	if len(req.AddDirs) != 0 {
		t.Fatalf("AddDirs should be empty; got %v", req.AddDirs)
	}
	h.restoreWorkspaceMCPConfig(state)
	if _, err := os.Stat(mcpPath); !os.IsNotExist(err) {
		t.Fatalf("expected mcp.json removed after restore; err=%v", err)
	}
}
