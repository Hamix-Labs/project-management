// Package git holds reconcile collaborators for the store git facade.
// SQL persistence stays in store; this package owns path/worktree diff helpers.
package git

import (
	"errors"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/model"
)

// PathKey delegates to gitwork.PathKey for Hamix ↔ git path compare.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func PathKey(path string) string {
	return gitwork.PathKey(path)
}

// MapGitworkBootstrapErr maps gitwork bootstrap errors to domain git API errors.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func MapGitworkBootstrapErr(err error) error {
	if errors.Is(err, gitwork.ErrBootstrapMismatch) {
		return domain.NewGitErr(domain.GitCodeBootstrapMismatch, "bootstrap path is not the same repository")
	}
	return err
}

// FilterLiveWorktrees drops prunable rows from git worktree list output.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FilterLiveWorktrees(live []gitwork.Worktree) []gitwork.Worktree {
	out := make([]gitwork.Worktree, 0, len(live))
	for _, wt := range live {
		if wt.Prunable {
			continue
		}
		out = append(out, wt)
	}
	return out
}

// LiveWorktreesByPath indexes live worktrees by normalized path.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func LiveWorktreesByPath(live []gitwork.Worktree) map[string]gitwork.Worktree {
	out := make(map[string]gitwork.Worktree, len(live))
	for _, wt := range live {
		out[PathKey(wt.Path)] = wt
	}
	return out
}

// LiveWorktreesByBranch indexes non-main live worktrees by branch name.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func LiveWorktreesByBranch(live []gitwork.Worktree) map[string]gitwork.Worktree {
	out := make(map[string]gitwork.Worktree, len(live))
	for _, wt := range live {
		if wt.IsMain || strings.TrimSpace(wt.Branch) == "" {
			continue
		}
		out[wt.Branch] = wt
	}
	return out
}

// CountBranchOwners counts registered worktrees bound to branchName.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func CountBranchOwners(rows []model.GitWorktree, branchName string, branchByID map[string]domain.GitBranch) int {
	n := 0
	for _, row := range rows {
		if row.IsMain {
			continue
		}
		br, ok := branchByID[row.BranchID]
		if ok && br.Name == branchName {
			n++
		}
	}
	return n
}
