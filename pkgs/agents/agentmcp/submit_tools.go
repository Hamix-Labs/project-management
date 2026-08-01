package agentmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	ToolSubmitCriteria = "hamix.submit_criteria_report"
	GroupReports       = "reports"
)

type submitCriteriaTool struct{}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (submitCriteriaTool) Name() string { return ToolSubmitCriteria }

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (submitCriteriaTool) Group() string { return GroupReports }

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (submitCriteriaTool) Description() string {
	return "Validate and write criteria-report.json for this execute phase. Required before execute finishes."
}

type submitCriteriaInput struct {
	Criteria []struct {
		ID          string `json:"id" jsonschema:"criterion id"`
		ClaimedDone bool   `json:"claimed_done" jsonschema:"whether the criterion is claimed done"`
		Evidence    string `json:"evidence" jsonschema:"evidence for the claim"`
	} `json:"criteria" jsonschema:"active criteria claims"`
}

type submitCriteriaOutput struct {
	OK      bool   `json:"ok"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (t submitCriteriaTool) Register(server *mcp.Server, sess *Session) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        t.Name(),
		Description: t.Description(),
	}, func(ctx context.Context, req *mcp.CallToolRequest, in submitCriteriaInput) (*mcp.CallToolResult, submitCriteriaOutput, error) {
		out, err := submitCriteria(sess, in)
		if err != nil {
			return toolErr(err), submitCriteriaOutput{}, nil
		}
		return nil, out, nil
	})
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func submitCriteria(sess *Session, in submitCriteriaInput) (submitCriteriaOutput, error) {
	if sess.Phase != PhaseExecute {
		return submitCriteriaOutput{}, fmt.Errorf("phase is %q; %s requires execute", sess.Phase, ToolSubmitCriteria)
	}
	entries := make([]sidecar.CriteriaEntry, 0, len(in.Criteria))
	seen := make(map[string]struct{}, len(in.Criteria))
	for _, c := range in.Criteria {
		id := strings.TrimSpace(c.ID)
		if id == "" {
			return submitCriteriaOutput{}, fmt.Errorf("empty criterion id")
		}
		if _, dup := seen[id]; dup {
			return submitCriteriaOutput{}, fmt.Errorf("duplicate criterion id %s", id)
		}
		seen[id] = struct{}{}
		if _, ok := sess.ActiveCriterionIDs[id]; !ok && len(sess.ActiveCriterionIDs) > 0 {
			return submitCriteriaOutput{}, fmt.Errorf("criterion %s is not active for this execute", id)
		}
		if len(c.Evidence) > sidecar.MaxFieldBytes() {
			return submitCriteriaOutput{}, fmt.Errorf("evidence too long for %s", id)
		}
		entries = append(entries, sidecar.CriteriaEntry{
			ID:          id,
			ClaimedDone: c.ClaimedDone,
			Evidence:    c.Evidence,
		})
	}
	for id := range sess.ActiveCriterionIDs {
		if _, ok := seen[id]; !ok {
			return submitCriteriaOutput{}, fmt.Errorf("missing active criterion %s", id)
		}
	}
	if err := sidecar.WriteCriteriaReport(sess.ReportDir, sess.CycleID, entries); err != nil {
		return submitCriteriaOutput{}, err
	}
	if err := writeReceipt(sess, ToolSubmitCriteria, PhaseExecute, sidecar.CriteriaSubmitReceiptPath(sess.ReportDir, sess.CycleID)); err != nil {
		return submitCriteriaOutput{}, err
	}
	path := sidecar.CriteriaReportPath(sess.ReportDir, sess.CycleID)
	return submitCriteriaOutput{OK: true, Path: path, Message: "criteria report submitted"}, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func writeReceipt(sess *Session, tool, phase, path string) error {
	return sidecar.WriteSubmitReceipt(path, sidecar.SubmitReceipt{
		Nonce:     sess.SubmitNonce,
		Phase:     phase,
		CycleID:   sess.CycleID,
		Tool:      tool,
		WrittenAt: time.Now().UTC().Format(time.RFC3339Nano),
	})
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func toolErr(err error) *mcp.CallToolResult {
	msg := err.Error()
	b, _ := json.Marshal(map[string]any{"ok": false, "error": msg})
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
	}
}
