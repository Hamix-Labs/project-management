package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// GitWriteStore covers git mutators that invoke gitwork.Service on disk.
type GitWriteStore interface {
	CreateGlobalGitRepository(ctx context.Context, input CreateGitRepositoryInput, gitSvc gitwork.Service) (domain.GitRepository, error)
	CreateGitRepository(ctx context.Context, projectID string, input CreateGitRepositoryInput, gitSvc gitwork.Service) (domain.GitRepository, error)
	CreateGitWorktreeForRepo(ctx context.Context, repoID string, input CreateGitWorktreeInput, gitSvc gitwork.Service) (domain.GitWorktree, error)
	CreateGitWorktree(ctx context.Context, projectID, repoID string, input CreateGitWorktreeInput, gitSvc gitwork.Service) (domain.GitWorktree, error)
	RemoveGitWorktreeFromDiskByID(ctx context.Context, worktreeID string, force bool, gitSvc gitwork.Service) error
	RemoveGitWorktreeFromDisk(ctx context.Context, projectID, worktreeID string, force bool, gitSvc gitwork.Service) error
	CreateGitBranch(ctx context.Context, projectID, repoID string, input CreateGitBranchInput, gitSvc gitwork.Service) (domain.GitBranch, error)
	DeleteGitBranch(ctx context.Context, projectID, branchID string, force bool, gitSvc gitwork.Service) error
	RepoWorktreeInventory(ctx context.Context, repo domain.GitRepository, gitSvc gitwork.Service) ([]WorktreeInventoryRow, error)
	RepoWorktreeCheckoutStatus(ctx context.Context, repo domain.GitRepository, gitSvc gitwork.Service) ([]WorktreeCheckoutStatusRow, error)
	ProbeGitWorktree(ctx context.Context, repoID, path string, gitSvc gitwork.Service) (GitWorktreeProbeResult, error)
	RegisterExistingGitWorktree(ctx context.Context, repoID, path, name string, bind BindBranchInput, gitSvc gitwork.Service) (domain.GitWorktree, error)
	ReconcileGitRepository(ctx context.Context, projectID, repoID string, input ReconcileGitInput, gitSvc gitwork.Service) (ReconcileGitOutput, error)
	RelocateGitRepository(ctx context.Context, projectID, repoID, path string, gitSvc gitwork.Service) (ReconcileGitOutput, error)
	RelocateGitWorktree(ctx context.Context, worktreeID, path string, gitSvc gitwork.Service) (domain.GitWorktree, error)
}
