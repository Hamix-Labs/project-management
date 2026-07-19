package gitinventory

import (
	"path/filepath"
	"strings"
	"unicode"
)

// ManagedWorktreePath returns the on-disk path for a Hamix-allocated worktree:
// `{dir(repoPath)}/.hamix/{repoID}/worktrees/{branchSlug}`.
//
//funclogmeasure:skip category=hot-path reason="Pure path helper without I/O; operation trace is emitted by the calling chokepoint."
func ManagedWorktreePath(repoPath, repoID, branch string) string {
	parent := filepath.Dir(filepath.Clean(strings.TrimSpace(repoPath)))
	id := strings.TrimSpace(repoID)
	slug := BranchPathSlug(branch)
	return filepath.ToSlash(filepath.Join(parent, ".hamix", id, "worktrees", slug))
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
