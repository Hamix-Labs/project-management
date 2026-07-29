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
	ToolSubmitVerify   = "hamix.submit_verify_report"
	GroupReports       = "reports"
)

type submitCriteriaTool struct{}

func (submitCriteriaTool) Name() string  { return ToolSubmitCriteria }
func (submitCriteriaTool) Group() string { return GroupReports }
func (submitCriteriaTool) Description() string {
	return "Validate and write criteria-report.json for this execute phase. Required before execute finishes."
}

type submitCriteriaInput struct {
	Criteria []struct {
		ID          string `json:"id" jsonschema:"criterion id"`
		ClaimedDone bool   `json:"claimed_done" jsonschema:"whether the criterion is claimed done"`
		Evidence    string `json:"evidence" jsonschema:"evidence for the claim"`
	} `json:"criteria" jsonschema:"active criteria claims"`
	Commits []struct {
		SHA    string `json:"sha" jsonschema:"commit sha"`
		Branch string `json:"branch,omitempty" jsonschema:"optional branch"`
	} `json:"commits,omitempty" jsonschema:"commits created in this execute visit"`
}

type submitCriteriaOutput struct {
	OK      bool   `json:"ok"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

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
	commits := make([]sidecar.CriteriaCommitClaim, 0, len(in.Commits))
	for _, c := range in.Commits {
		sha := strings.TrimSpace(c.SHA)
		if sha == "" {
			return submitCriteriaOutput{}, fmt.Errorf("empty commit sha")
		}
		commits = append(commits, sidecar.CriteriaCommitClaim{
			SHA:    sha,
			Branch: strings.TrimSpace(c.Branch),
		})
	}
	if err := sidecar.WriteCriteriaReport(sess.ReportDir, sess.CycleID, entries, commits); err != nil {
		return submitCriteriaOutput{}, err
	}
	if err := writeReceipt(sess, ToolSubmitCriteria, PhaseExecute, sidecar.CriteriaSubmitReceiptPath(sess.ReportDir, sess.CycleID)); err != nil {
		return submitCriteriaOutput{}, err
	}
	path := sidecar.CriteriaReportPath(sess.ReportDir, sess.CycleID)
	return submitCriteriaOutput{OK: true, Path: path, Message: "criteria report submitted"}, nil
}

type submitVerifyTool struct{}

func (submitVerifyTool) Name() string  { return ToolSubmitVerify }
func (submitVerifyTool) Group() string { return GroupReports }
func (submitVerifyTool) Description() string {
	return "Validate and write verify-report.json for this verify phase. Required before verify finishes."
}

type submitVerifyInput struct {
	Criteria []struct {
		ID        string `json:"id" jsonschema:"criterion id"`
		Verified  bool   `json:"verified" jsonschema:"whether the criterion passed verify"`
		Reasoning string `json:"reasoning" jsonschema:"reasoning for the verdict"`
	} `json:"criteria" jsonschema:"active criteria verdicts"`
}

type submitVerifyOutput struct {
	OK      bool   `json:"ok"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (t submitVerifyTool) Register(server *mcp.Server, sess *Session) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        t.Name(),
		Description: t.Description(),
	}, func(ctx context.Context, req *mcp.CallToolRequest, in submitVerifyInput) (*mcp.CallToolResult, submitVerifyOutput, error) {
		out, err := submitVerify(sess, in)
		if err != nil {
			return toolErr(err), submitVerifyOutput{}, nil
		}
		return nil, out, nil
	})
}

func submitVerify(sess *Session, in submitVerifyInput) (submitVerifyOutput, error) {
	if sess.Phase != PhaseVerify {
		return submitVerifyOutput{}, fmt.Errorf("phase is %q; %s requires verify", sess.Phase, ToolSubmitVerify)
	}
	entries := make([]sidecar.VerifyEntry, 0, len(in.Criteria))
	seen := make(map[string]struct{}, len(in.Criteria))
	for _, c := range in.Criteria {
		id := strings.TrimSpace(c.ID)
		if id == "" {
			return submitVerifyOutput{}, fmt.Errorf("empty criterion id")
		}
		if _, dup := seen[id]; dup {
			return submitVerifyOutput{}, fmt.Errorf("duplicate criterion id %s", id)
		}
		seen[id] = struct{}{}
		if _, ok := sess.ActiveCriterionIDs[id]; !ok && len(sess.ActiveCriterionIDs) > 0 {
			return submitVerifyOutput{}, fmt.Errorf("criterion %s is not active for this verify", id)
		}
		if c.Verified && len(strings.TrimSpace(c.Reasoning)) < sidecar.MinVerifyReasoningChars() {
			return submitVerifyOutput{}, fmt.Errorf("reasoning too short for verified criterion %s", id)
		}
		if len(c.Reasoning) > sidecar.MaxFieldBytes() {
			return submitVerifyOutput{}, fmt.Errorf("reasoning too long for %s", id)
		}
		entries = append(entries, sidecar.VerifyEntry{
			ID:        id,
			Verified:  c.Verified,
			Reasoning: c.Reasoning,
		})
	}
	for id := range sess.ActiveCriterionIDs {
		if _, ok := seen[id]; !ok {
			return submitVerifyOutput{}, fmt.Errorf("missing active criterion %s", id)
		}
	}
	if err := sidecar.WriteVerifyReport(sess.ReportDir, sess.CycleID, entries); err != nil {
		return submitVerifyOutput{}, err
	}
	if err := writeReceipt(sess, ToolSubmitVerify, PhaseVerify, sidecar.VerifySubmitReceiptPath(sess.ReportDir, sess.CycleID)); err != nil {
		return submitVerifyOutput{}, err
	}
	path := sidecar.VerifyReportPath(sess.ReportDir, sess.CycleID)
	return submitVerifyOutput{OK: true, Path: path, Message: "verify report submitted"}, nil
}

func writeReceipt(sess *Session, tool, phase, path string) error {
	return sidecar.WriteSubmitReceipt(path, sidecar.SubmitReceipt{
		Nonce:     sess.SubmitNonce,
		Phase:     phase,
		CycleID:   sess.CycleID,
		Tool:      tool,
		WrittenAt: time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func toolErr(err error) *mcp.CallToolResult {
	msg := err.Error()
	b, _ := json.Marshal(map[string]any{"ok": false, "error": msg})
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
	}
}
