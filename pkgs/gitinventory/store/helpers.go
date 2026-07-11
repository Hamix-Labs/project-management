package store

import "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/internal/git"

// worktreePathKey normalizes paths for Hamix ↔ git compare (delegates to internal/git).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func worktreePathKey(path string) string {
	return git.PathKey(path)
}
