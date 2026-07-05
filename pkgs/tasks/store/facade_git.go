package store

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/internal/git"
)

// facade_git.go documents git facade entrypoints. Reconcile orchestration lives
// in reconcile_git.go; path/worktree diff helpers live in internal/git/.

// worktreePathKey normalizes paths for Hamix ↔ git compare (delegates to internal/git).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func worktreePathKey(path string) string {
	return git.PathKey(path)
}
