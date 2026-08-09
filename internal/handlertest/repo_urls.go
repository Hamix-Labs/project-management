package handlertest

import "net/url"

// RepoPathWithWorktree builds GET /repo/file query for handler contract tests.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only URL builder; not part of production trace paths."
func RepoPathWithWorktree(worktreeID, path string) string {
	q := url.Values{}
	q.Set("worktree_id", worktreeID)
	if path != "" {
		q.Set("path", path)
	}
	return "/repo/file?" + q.Encode()
}

// RepoSearchWithWorktree builds GET /repo/search query for handler contract tests.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only URL builder; not part of production trace paths."
func RepoSearchWithWorktree(worktreeID, q string) string {
	v := url.Values{}
	v.Set("worktree_id", worktreeID)
	v.Set("q", q)
	return "/repo/search?" + v.Encode()
}

// RepoFilesWithWorktree builds GET /repo/files query for handler contract tests.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only URL builder; not part of production trace paths."
func RepoFilesWithWorktree(worktreeID string) string {
	v := url.Values{}
	v.Set("worktree_id", worktreeID)
	return "/repo/files?" + v.Encode()
}

// RepoValidateRangeWithWorktree builds GET /repo/validate-range query.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only URL builder; not part of production trace paths."
func RepoValidateRangeWithWorktree(worktreeID, path, start, end string) string {
	v := url.Values{}
	v.Set("worktree_id", worktreeID)
	v.Set("path", path)
	v.Set("start", start)
	v.Set("end", end)
	return "/repo/validate-range?" + v.Encode()
}

// RepoDiffWithWorktree builds GET /repo/diff query.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only URL builder; not part of production trace paths."
func RepoDiffWithWorktree(worktreeID, sha string) string {
	v := url.Values{}
	v.Set("worktree_id", worktreeID)
	v.Set("sha", sha)
	return "/repo/diff?" + v.Encode()
}
