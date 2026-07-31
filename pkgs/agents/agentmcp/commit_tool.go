package agentmcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	ToolCommit   = "hamix.commit"
	GroupGit     = "git"
	maxCommitMsg = 16 * 1024
)

type commitTool struct{}

//funclogmeasure:skip category=hot-path reason="MCP tool metadata accessor."
func (commitTool) Name() string { return ToolCommit }

//funclogmeasure:skip category=hot-path reason="MCP tool metadata accessor."
func (commitTool) Group() string { return GroupGit }

//funclogmeasure:skip category=hot-path reason="MCP tool metadata accessor."
func (commitTool) Description() string {
	return "Commit the current git index in the task worktree and record the resulting SHA in the cycle commit register. Stage files with Shell git add first. Do not use Shell git commit."
}

type commitInput struct {
	Message string `json:"message" jsonschema:"commit message (required)"`
}

type commitOutput struct {
	OK      bool   `json:"ok"`
	SHA     string `json:"sha"`
	Message string `json:"message"`
}

//funclogmeasure:skip category=hot-path reason="MCP SDK registration; business logic is in runCommit."
func (t commitTool) Register(server *mcp.Server, sess *Session) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        t.Name(),
		Description: t.Description(),
	}, func(ctx context.Context, req *mcp.CallToolRequest, in commitInput) (*mcp.CallToolResult, commitOutput, error) {
		out, err := runCommit(ctx, sess, in)
		if err != nil {
			return toolErr(err), commitOutput{}, nil
		}
		return nil, out, nil
	})
}

//funclogmeasure:skip category=hot-path reason="Git + register I/O; callers surface errors to the agent."
func runCommit(ctx context.Context, sess *Session, in commitInput) (commitOutput, error) {
	if sess.Phase != PhaseExecute {
		return commitOutput{}, fmt.Errorf("phase is %q; %s requires execute", sess.Phase, ToolCommit)
	}
	workDir := strings.TrimSpace(sess.WorkingDir)
	if workDir == "" {
		return commitOutput{}, fmt.Errorf("working_dir is empty")
	}
	msg := strings.TrimSpace(in.Message)
	if msg == "" {
		return commitOutput{}, fmt.Errorf("message is required")
	}
	if len(msg) > maxCommitMsg {
		return commitOutput{}, fmt.Errorf("message too long")
	}

	msgFile, err := os.CreateTemp("", "hamix-commit-msg-*.txt")
	if err != nil {
		return commitOutput{}, fmt.Errorf("create message file: %w", err)
	}
	msgPath := msgFile.Name()
	defer os.Remove(msgPath)
	if _, err := msgFile.WriteString(msg); err != nil {
		_ = msgFile.Close()
		return commitOutput{}, fmt.Errorf("write message file: %w", err)
	}
	if err := msgFile.Close(); err != nil {
		return commitOutput{}, fmt.Errorf("close message file: %w", err)
	}

	commitOut, err := gitRun(ctx, workDir, "commit", "-F", msgPath)
	if err != nil {
		detail := strings.TrimSpace(commitOut)
		if detail == "" {
			detail = err.Error()
		}
		return commitOutput{}, fmt.Errorf("git commit failed: %s", detail)
	}

	shaOut, err := gitRun(ctx, workDir, "rev-parse", "HEAD")
	if err != nil {
		return commitOutput{}, fmt.Errorf("rev-parse HEAD after commit: %w", err)
	}
	sha := strings.TrimSpace(shaOut)
	if sha == "" {
		return commitOutput{}, fmt.Errorf("empty HEAD after commit")
	}

	branch := ""
	if bOut, bErr := gitRun(ctx, workDir, "branch", "--show-current"); bErr == nil {
		branch = strings.TrimSpace(bOut)
	}

	entry := sidecar.CommitRegisterEntry{
		SHA:       sha,
		Message:   msg,
		Branch:    branch,
		WrittenAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := sidecar.AppendCommitRegister(sess.ReportDir, sess.CycleID, entry); err != nil {
		if retryErr := sidecar.AppendCommitRegister(sess.ReportDir, sess.CycleID, entry); retryErr != nil {
			return commitOutput{}, fmt.Errorf(
				"commit created at %s but register append failed: %v", sha, retryErr)
		}
	}

	return commitOutput{OK: true, SHA: sha, Message: msg}, nil
}

//funclogmeasure:skip category=hot-path reason="Thin git exec wrapper."
func gitRun(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
