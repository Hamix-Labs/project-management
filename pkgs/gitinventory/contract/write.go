package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
)

// GitWriteStore covers git mutators that invoke the store-injected gitwork.Service on disk.
type GitWriteStore interface {
	CreateGlobalGitRepository(ctx context.Context, input CreateGitRepositoryInput) (domain.GitRepository, error)
	CreateGitRepository(ctx context.Context, projectID string, input CreateGitRepositoryInput) (domain.GitRepository, error)
	CreateGitWorktreeForRepo(ctx context.Context, repoID string, input CreateGitWorktreeInput) (domain.GitWorktree, error)
	CreateGitWorktree(ctx context.Context, projectID, repoID string, input CreateGitWorktreeInput) (domain.GitWorktree, error)
	// AllocateTaskWorktree fetches origin and creates a Hamix-managed worktree+branch for a task.
	AllocateTaskWorktree(ctx context.Context, repoID, taskID string) (domain.GitWorktree, error)
	RemoveGitWorktreeFromDiskByID(ctx context.Context, worktreeID string, force bool) error
	RemoveGitWorktreeFromDisk(ctx context.Context, projectID, worktreeID string, force bool) error
	CreateGitBranch(ctx context.Context, projectID, repoID string, input CreateGitBranchInput) (domain.GitBranch, error)
	DeleteGitBranch(ctx context.Context, projectID, branchID string, force bool) error
	RepoWorktreeInventory(ctx context.Context, repo domain.GitRepository) ([]WorktreeInventoryRow, error)
	RepoWorktreeCheckoutStatus(ctx context.Context, repo domain.GitRepository) ([]WorktreeCheckoutStatusRow, error)
	ProbeGitWorktree(ctx context.Context, repoID, path string) (GitWorktreeProbeResult, error)
	RegisterExistingGitWorktree(ctx context.Context, repoID, path, name string, bind BindBranchInput) (domain.GitWorktree, error)
	ReconcileGitRepository(ctx context.Context, projectID, repoID string, input ReconcileGitInput) (ReconcileGitOutput, error)
	// SyncGitRepository fetches origin and refreshes metadata without discover.
	SyncGitRepository(ctx context.Context, repoID string) (ReconcileGitOutput, error)
	RelocateGitRepository(ctx context.Context, projectID, repoID, path string) (ReconcileGitOutput, error)
	RelocateGitWorktree(ctx context.Context, worktreeID, path string) (domain.GitWorktree, error)
}
