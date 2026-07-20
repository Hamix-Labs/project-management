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
	UnregisterGitWorktreeByID(ctx context.Context, worktreeID string) error
	UnregisterGitWorktree(ctx context.Context, projectID, worktreeID string) error
	ListGitBranchesByRepo(ctx context.Context, repoID string) ([]domain.GitBranch, error)
	ListGitBranches(ctx context.Context, projectID, repoID string) ([]domain.GitBranch, error)
	// WorktreeStaleMap reports stale managed worktrees for a repository.
	WorktreeStaleMap(ctx context.Context, repoID string, now time.Time) (map[string]bool, error)
}
