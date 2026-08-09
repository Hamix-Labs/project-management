package gitexec

import (
	"context"
	"errors"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitcore"
)

// ErrNoWorkTree reports that dir has no checked-out work tree to list, either
// because it is not a git repository at all or because it is a bare one.
var ErrNoWorkTree = errors.New("gitexec: no git work tree")

// ListFiles returns work-tree-relative paths for every tracked and untracked
// file in dir, with slash separators.
//
// `--cached --others --exclude-standard` is what makes this gitignore-aware:
// git applies nested .gitignore files, .git/info/exclude, and core.excludesFile
// itself, so callers need no exclusion list of their own. Tracked files are
// listed even when a rule would otherwise ignore them, matching what a user
// expects to be able to reference.
//
// limit caps the result; truncated reports whether it applied. A dir that is
// not a git work tree returns ErrNoWorkTree so callers can fall back.
//
//funclogmeasure:skip category=hot-path reason="Pure subprocess helper; operation trace is emitted by the calling chokepoint."
func ListFiles(ctx context.Context, dir string, limit int) (paths []string, truncated bool, err error) {
	out, err := Run(ctx, dir, "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	if err != nil {
		if isNoWorkTreeErr(err) {
			return nil, false, ErrNoWorkTree
		}
		return nil, false, err
	}
	// -z separates records with NUL so paths containing newlines or quotes
	// survive intact; the default output would quote and escape them.
	fields := strings.Split(out, "\x00")
	paths = make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		if limit > 0 && len(paths) >= limit {
			return paths, true, nil
		}
		paths = append(paths, field)
	}
	return paths, false, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func isNoWorkTreeErr(err error) bool {
	return gitcore.IsNotARepository(err) ||
		gitcore.StderrContains(err, "must be run in a work tree")
}
