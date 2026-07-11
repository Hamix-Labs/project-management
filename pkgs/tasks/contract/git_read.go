package contract

import gitcontract "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"

// GitReadStore covers git entity reads and DB-only mutations (no gitSvc).
type GitReadStore = gitcontract.GitReadStore

// GitWriteStore covers git mutators that invoke gitwork.Service on disk.
type GitWriteStore = gitcontract.GitWriteStore

// GitRepositoryListSummary augments a repository row with list-page metadata.
type GitRepositoryListSummary = gitcontract.GitRepositoryListSummary

// CreateGitRepositoryInput registers a main git checkout for a project.
type CreateGitRepositoryInput = gitcontract.CreateGitRepositoryInput

// CreateGitWorktreeInput adds a worktree on disk and persists the row.
type CreateGitWorktreeInput = gitcontract.CreateGitWorktreeInput

// CreateGitBranchInput creates a local branch via git.
type CreateGitBranchInput = gitcontract.CreateGitBranchInput

// BindBranchInput registers or creates a repo-level branch row for worktree assignment.
type BindBranchInput = gitcontract.BindBranchInput

// WorktreeInventoryRow is a live git worktree plus Hamix registration state.
type WorktreeInventoryRow = gitcontract.WorktreeInventoryRow

// GitWorktreeProbeResult describes whether a path is a linked, registerable worktree.
type GitWorktreeProbeResult = gitcontract.GitWorktreeProbeResult

// WorktreeCheckoutStatusRow is live checkout git state for one registered worktree.
type WorktreeCheckoutStatusRow = gitcontract.WorktreeCheckoutStatusRow

// ReconcileGitInput configures repository/worktree path sync with git.
type ReconcileGitInput = gitcontract.ReconcileGitInput

// ReconcileGitOutput is the structured reconcile result for API and operators.
type ReconcileGitOutput = gitcontract.ReconcileGitOutput

// ReconcileReport counts reconcile actions and skipped rows.
type ReconcileReport = gitcontract.ReconcileReport

// ReconcileSkippedWorktree describes a DB row reconcile could not remove.
type ReconcileSkippedWorktree = gitcontract.ReconcileSkippedWorktree

// ReconcileNeedsBranchBind describes a live worktree without Hamix branch binding.
type ReconcileNeedsBranchBind = gitcontract.ReconcileNeedsBranchBind
