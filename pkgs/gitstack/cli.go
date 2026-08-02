// Package gitstack wraps the GitHub CLI `gh stack` extension for Hamix-managed
// worktree families (ADR-0097).
package gitstack

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CLI runs non-interactive gh stack commands in a worktree directory.
type CLI interface {
	Init(ctx context.Context, workDir, baseBranch, firstBranch string) error
	Add(ctx context.Context, workDir, branch string) error
	Submit(ctx context.Context, workDir string) (string, error)
	Rebase(ctx context.Context, workDir string) error
	ViewJSON(ctx context.Context, workDir string) (string, error)
}

// RealCLI invokes the `gh` binary with GH_PROMPT_DISABLED=1.
type RealCLI struct{}

// New returns the default CLI implementation.
//
//funclogmeasure:skip category=hot-path reason="Pure constructor without I/O."
func New() CLI { return RealCLI{} }

// Nop is a no-op CLI for tests that do not exercise gh stack.
type Nop struct{}

//funclogmeasure:skip category=hot-path reason="Test double no-op."
func (Nop) Init(context.Context, string, string, string) error { return nil }

//funclogmeasure:skip category=hot-path reason="Test double no-op."
func (Nop) Add(context.Context, string, string) error { return nil }

//funclogmeasure:skip category=hot-path reason="Test double no-op."
func (Nop) Submit(context.Context, string) (string, error) { return "", nil }

//funclogmeasure:skip category=hot-path reason="Test double no-op."
func (Nop) Rebase(context.Context, string) error { return nil }

//funclogmeasure:skip category=hot-path reason="Test double no-op."
func (Nop) ViewJSON(context.Context, string) (string, error) { return "[]", nil }

//funclogmeasure:skip category=hot-path reason="Thin gh exec wrapper; callers surface errors."
func (RealCLI) Init(ctx context.Context, workDir, baseBranch, firstBranch string) error {
	baseBranch = strings.TrimSpace(baseBranch)
	firstBranch = strings.TrimSpace(firstBranch)
	if firstBranch == "" {
		return fmt.Errorf("gitstack: first branch required")
	}
	args := []string{"stack", "init"}
	if baseBranch != "" {
		args = append(args, "--base", baseBranch)
	}
	args = append(args, firstBranch)
	_, err := run(ctx, workDir, args...)
	return err
}

//funclogmeasure:skip category=hot-path reason="Thin gh exec wrapper; callers surface errors."
func (RealCLI) Add(ctx context.Context, workDir, branch string) error {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return fmt.Errorf("gitstack: branch required")
	}
	_, err := run(ctx, workDir, "stack", "add", branch)
	return err
}

//funclogmeasure:skip category=hot-path reason="Thin gh exec wrapper; callers surface errors."
func (RealCLI) Submit(ctx context.Context, workDir string) (string, error) {
	return run(ctx, workDir, "stack", "submit", "--auto", "--open")
}

//funclogmeasure:skip category=hot-path reason="Thin gh exec wrapper; callers surface errors."
func (RealCLI) Rebase(ctx context.Context, workDir string) error {
	_, err := run(ctx, workDir, "stack", "rebase", "--no-trunk")
	return err
}

//funclogmeasure:skip category=hot-path reason="Thin gh exec wrapper; callers surface errors."
func (RealCLI) ViewJSON(ctx context.Context, workDir string) (string, error) {
	return run(ctx, workDir, "stack", "view", "--json")
}

//funclogmeasure:skip category=hot-path reason="Thin gh exec wrapper."
func run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GH_PROMPT_DISABLED=1", "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	text := string(out)
	if err != nil {
		detail := strings.TrimSpace(text)
		if detail == "" {
			detail = err.Error()
		}
		return text, fmt.Errorf("gh %s: %s", strings.Join(args, " "), detail)
	}
	return text, nil
}
