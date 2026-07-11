package contract

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
)

// GitReadStore covers git entity reads and DB-only mutations (no gitSvc).
type GitReadStore interface {
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
}
