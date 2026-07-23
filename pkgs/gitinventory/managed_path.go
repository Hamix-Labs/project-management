package gitinventory

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// EnvManagedWorktreeRoot overrides the default managed-worktree root when set.
const EnvManagedWorktreeRoot = "HAMIX_MANAGED_WORKTREE_ROOT"

// ManagedWorktreeRoot returns the root directory for Hamix-allocated worktrees:
// `$HAMIX_MANAGED_WORKTREE_ROOT` when set, otherwise `{UserConfigDir}/hamix`
// (falling back to `{UserHomeDir}/hamix` if UserConfigDir is unavailable).
//
//funclogmeasure:skip category=hot-path reason="Pure path helper without I/O beyond env/UserConfigDir; operation trace is emitted by the calling chokepoint."
func ManagedWorktreeRoot() string {
	if v := strings.TrimSpace(os.Getenv(EnvManagedWorktreeRoot)); v != "" {
		return filepath.Clean(v)
	}
	base, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(base) == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil || strings.TrimSpace(home) == "" {
			return "hamix"
		}
		base = home
	}
	return filepath.Join(base, "hamix")
}

// ManagedWorktreePath returns the on-disk path for a Hamix-allocated worktree:
// `{ManagedWorktreeRoot}/worktrees/{repoID}/{branchSlug}`.
//
// repoPath is retained for call-site compatibility and is not used for the root.
//
//funclogmeasure:skip category=hot-path reason="Pure path helper without I/O; operation trace is emitted by the calling chokepoint."
func ManagedWorktreePath(repoPath, repoID, branch string) string {
	_ = repoPath
	id := strings.TrimSpace(repoID)
	slug := BranchPathSlug(branch)
	return filepath.ToSlash(filepath.Join(ManagedWorktreeRoot(), "worktrees", id, slug))
}

// BranchPathSlug turns a git branch name into a single filesystem path segment.
//
//funclogmeasure:skip category=hot-path reason="Pure path helper without I/O; operation trace is emitted by the calling chokepoint."
func BranchPathSlug(branch string) string {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return "branch"
	}
	var b strings.Builder
	b.Grow(len(branch))
	prevDash := false
	for _, r := range branch {
		switch {
		case r == '/' || r == '\\':
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' || r == '-':
			b.WriteRune(r)
			prevDash = r == '-'
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "branch"
	}
	return out
}
