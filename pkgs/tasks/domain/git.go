package domain

import (
	"errors"
	"time"
)

// Git API error codes returned in JSON {"error","code"} responses.
const (
	GitCodeNotARepository     = "not_a_git_repository"
	GitCodePathExists         = "path_exists"
	GitCodeBranchExists       = "branch_exists"
	GitCodeBranchCheckedOut   = "branch_checked_out"
	GitCodeHasRunningTask     = "has_running_task"
	GitCodeRepositoryNotFound = "repository_not_found"
	GitCodeWorktreeNotFound   = "worktree_not_found"
	GitCodeBranchNotFound     = "branch_not_found"
	GitCodeDuplicate          = "duplicate"
	// GitCodeBranchBoundToWorktree is returned when a branch is already bound
	// to a different worktree row in the fixed worktree-branch model.
	GitCodeBranchBoundToWorktree = "branch_bound_to_worktree"
	// GitCodeBranchActiveElsewhere is returned when a branch is the active
	// checkout in another worktree and Hamix rejects binding/running against
	// it (replaces the soft "checked out elsewhere" warning). See ADR-0037.
	GitCodeBranchActiveElsewhere = "branch_active_elsewhere"
	// GitCodeBranchNotAssociated is returned when a task binds a (worktree,
	// branch) pair that has no worktree_branches association row.
	GitCodeBranchNotAssociated = "branch_not_associated"
	// GitCodeProjectRepoMismatch is returned when a task's project belongs to
	// a different repository than its bound worktree.
	GitCodeProjectRepoMismatch = "project_repo_mismatch"
	// GitCodeBootstrapMismatch is returned when a bootstrap/relocate path is
	// not the same git repository as the registered row.
	GitCodeBootstrapMismatch = "bootstrap_mismatch"
)

// GitErr is a domain error with a stable API code for git entity routes.
type GitErr struct {
	Code string
	Msg  string
}

func (e *GitErr) Error() string { return e.Msg }

// NewGitErr returns an error tagged with a stable git API code.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func NewGitErr(code, msg string) error {
	return &GitErr{Code: code, Msg: msg}
}

// GitErrCode returns the stable code when err wraps *GitErr.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func GitErrCode(err error) string {
	var ge *GitErr
	if errors.As(err, &ge) {
		return ge.Code
	}
	return ""
}

// GitRepository is a registered main git checkout. Globally unique on Path
// (one row per canonical path, shared across projects). See ADR-0037.
type GitRepository struct {
	ID            string    `json:"id"`
	Path          string    `json:"path"`
	GitCommonDir  string    `json:"git_common_dir"`
	HostPath      string    `json:"host_path"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GitWorktree is a linked working directory for a GitRepository.
//
// BranchID binds the worktree to a repo-level branch row (fixed worktree-branch
// model). Plain indexed column — set at registration; see migrate contract.
type GitWorktree struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repository_id"`
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	IsMain       bool      `json:"is_main"`
	BranchID     string    `json:"branch_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// GitBranch is a local branch tracked for a GitRepository (repo-level ref).
type GitBranch struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repository_id"`
	Name         string    `json:"name"`
	HeadSHA      string    `json:"head_sha"`
	CreatedAt    time.Time `json:"created_at"`
}
