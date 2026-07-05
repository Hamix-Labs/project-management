package contract

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// GitRepositoryListSummary augments a repository row with list-page metadata.
type GitRepositoryListSummary struct {
	Repository          domain.GitRepository
	MainBranchName      string
	LinkedWorktreeCount int
}

// CreateGitRepositoryInput registers a main git checkout for a project.
type CreateGitRepositoryInput struct {
	Path          string
	HostPath      string
	DefaultBranch string
}

// CreateGitWorktreeInput adds a worktree on disk and persists the row.
type CreateGitWorktreeInput struct {
	Path         string
	Name         string
	Branch       string
	CreateBranch bool
	StartPoint   string
	ForceRemove  bool
}

// CreateGitBranchInput creates a local branch via git.
type CreateGitBranchInput struct {
	Name       string
	StartPoint string
}

// BindBranchInput registers or creates a repo-level branch row for worktree assignment.
type BindBranchInput struct {
	Name         string
	CreateBranch bool
	StartPoint   string
}

// WorktreeInventoryRow is a live git worktree plus Hamix registration state.
type WorktreeInventoryRow struct {
	Path       string
	Branch     string
	IsMain     bool
	Detached   bool
	Registered bool
	Locked     bool
	Prunable   bool
}

// GitWorktreeProbeResult describes whether a path is a linked, registerable worktree.
type GitWorktreeProbeResult struct {
	Path       string
	Linked     bool
	IsMain     bool
	Branch     string
	Registered bool
}

// WorktreeCheckoutStatusRow is live checkout git state for one registered worktree.
type WorktreeCheckoutStatusRow struct {
	WorktreeID string
	Available  bool
	Reason     string // path_missing | git_error
	Status     gitwork.CheckoutStatus
}

// ReconcileGitInput configures repository/worktree path sync with git.
type ReconcileGitInput struct {
	BootstrapPath         string
	RepairGit             bool
	DryRun                bool
	AllowRemove           bool
	AllowCheckoutDiscover bool
	AllowDiscover         bool
}

// ReconcileGitOutput is the structured reconcile result for API and operators.
type ReconcileGitOutput struct {
	Status string
	Report ReconcileReport
}

// ReconcileReport counts reconcile actions and skipped rows.
type ReconcileReport struct {
	RepoPathUpdated      bool
	WorktreesPathUpdated int
	WorktreesAdded       int
	WorktreesRemoved     int
	BranchesHeadUpdated  int
	ResolutionSource     string
	DiscoveredPath       string
	WorktreesSkipped     []ReconcileSkippedWorktree
	NeedsBranchBind      []ReconcileNeedsBranchBind
}

// ReconcileSkippedWorktree describes a DB row reconcile could not remove.
type ReconcileSkippedWorktree struct {
	WorktreeID string
	Reason     string
}

// ReconcileNeedsBranchBind describes a live worktree without Hamix branch binding.
type ReconcileNeedsBranchBind struct {
	Path   string
	Branch string
}
