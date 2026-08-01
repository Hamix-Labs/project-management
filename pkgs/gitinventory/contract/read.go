package contract

import (
	"context"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
)

// GitInventoryStore covers git entity inventory: reads plus DB-only mutations (no gitSvc).
type GitInventoryStore interface {
	ListAllGitRepositories(ctx context.Context) ([]domain.GitRepository, error)
	ListAllGitRepositoriesWithSummary(ctx context.Context) ([]GitRepositoryListSummary, error)
	ListGitRepositories(ctx context.Context, projectID string) ([]domain.GitRepository, error)
	GetGitRepositoryByID(ctx context.Context, repoID string) (domain.GitRepository, error)
	GetGitRepository(ctx context.Context, projectID, repoID string) (domain.GitRepository, error)
	DeleteGlobalGitRepository(ctx context.Context, repoID string) error
	DeleteGitRepository(ctx context.Context, projectID, repoID string) error
	ListGitWorktreesByRepo(ctx context.Context, repoID string) ([]domain.GitWorktree, error)
	ListGitWorktrees(ctx context.Context, projectID, repoID string) ([]domain.GitWorktree, error)
	GetGitWorktreeByID(ctx context.Context, worktreeID string) (domain.GitWorktree, error)
	UnregisterGitWorktreeByID(ctx context.Context, worktreeID string) error
	UnregisterGitWorktree(ctx context.Context, projectID, worktreeID string) error
	ListGitBranchesByRepo(ctx context.Context, repoID string) ([]domain.GitBranch, error)
	ListGitBranches(ctx context.Context, projectID, repoID string) ([]domain.GitBranch, error)
	GetGitBranchByID(ctx context.Context, branchID string) (domain.GitBranch, error)
	// WorktreeStaleMap reports stale flags for the given worktrees.
	// Callers that already listed worktrees should pass those rows (no re-list).
	WorktreeStaleMap(ctx context.Context, worktrees []domain.GitWorktree, now time.Time) (map[string]bool, error)
}
