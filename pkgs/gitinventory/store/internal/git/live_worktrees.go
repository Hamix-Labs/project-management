// Package git holds reconcile collaborators for the store git facade.
// SQL persistence stays in store; this package owns path/worktree diff helpers.
package git

import (
	"errors"
	"os"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
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

// FilterLiveWorktrees keeps worktrees that are live for reconcile: not
// prunable and whose Path still exists on disk. Missing paths (os.IsNotExist)
// are treated as not live so vanished linked worktrees can be removed.
//
//funclogmeasure:skip category=hot-path reason="Stat filter for reconcile; operation trace is emitted by the calling chokepoint."
func FilterLiveWorktrees(live []gitwork.Worktree) []gitwork.Worktree {
	out := make([]gitwork.Worktree, 0, len(live))
	for _, wt := range live {
		if wt.Prunable {
			continue
		}
		if _, err := os.Stat(wt.Path); os.IsNotExist(err) {
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
