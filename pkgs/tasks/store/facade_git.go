package store

import (
	"context"
	"log/slog"

	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

// Git store input and result aliases — re-exported for handlers and harness callers.
type (
	CreateGitRepositoryInput     = gitinventorystore.CreateGitRepositoryInput
	CreateGitWorktreeInput       = gitinventorystore.CreateGitWorktreeInput
	CreateGitBranchInput         = gitinventorystore.CreateGitBranchInput
	BindBranchInput              = gitinventorystore.BindBranchInput
	GitRepositoryListSummary     = gitinventorystore.GitRepositoryListSummary
	WorktreeInventoryRow         = gitinventorystore.WorktreeInventoryRow
	GitWorktreeProbeResult       = gitinventorystore.GitWorktreeProbeResult
	WorktreeCheckoutStatusRow    = gitinventorystore.WorktreeCheckoutStatusRow
	ReconcileGitInput            = gitinventorystore.ReconcileGitInput
	ReconcileGitOutput           = gitinventorystore.ReconcileGitOutput
	ReconcileReport              = gitinventorystore.ReconcileReport
	ReconcileSkippedWorktree     = gitinventorystore.ReconcileSkippedWorktree
	ReconcileNeedsBranchBind     = gitinventorystore.ReconcileNeedsBranchBind
)

// FindWorktreeInInventory delegates to the git inventory store helper.
var FindWorktreeInInventory = gitinventorystore.FindWorktreeInInventory

// ListAllGitRepositories returns every registered repository ordered by created_at.
func (s *Store) ListAllGitRepositories(ctx context.Context) ([]gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListAllGitRepositories")
	return s.git.ListAllGitRepositories(ctx)
}

// ListAllGitRepositoriesWithSummary returns repositories with list-page metadata.
func (s *Store) ListAllGitRepositoriesWithSummary(ctx context.Context) ([]GitRepositoryListSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListAllGitRepositoriesWithSummary")
	return s.git.ListAllGitRepositoriesWithSummary(ctx)
}

// CreateGlobalGitRepository registers a repository without project scoping.
func (s *Store) CreateGlobalGitRepository(ctx context.Context, input CreateGitRepositoryInput, gitSvc gitwork.Service) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGlobalGitRepository")
	return s.git.CreateGlobalGitRepository(ctx, input, gitSvc)
}

// DeleteGlobalGitRepository removes a repository by ID.
func (s *Store) DeleteGlobalGitRepository(ctx context.Context, repoID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteGlobalGitRepository")
	return s.git.DeleteGlobalGitRepository(ctx, repoID)
}

// ListGitRepositories returns all registered repositories ordered by created_at.
func (s *Store) ListGitRepositories(ctx context.Context, projectID string) ([]gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListGitRepositories")
	return s.git.ListGitRepositories(ctx, projectID)
}

// CountGitRepositories returns the total number of registered git repositories.
func (s *Store) CountGitRepositories(ctx context.Context) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CountGitRepositories")
	return s.git.CountGitRepositories(ctx)
}

// GetGitRepository returns one repository by ID.
func (s *Store) GetGitRepository(ctx context.Context, projectID, repoID string) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitRepository")
	return s.git.GetGitRepository(ctx, projectID, repoID)
}

// GetGitRepositoryByID loads a repository row by primary key.
func (s *Store) GetGitRepositoryByID(ctx context.Context, repoID string) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitRepositoryByID")
	return s.git.GetGitRepositoryByID(ctx, repoID)
}

// CreateGitRepository validates path with git, then inserts repository + main worktree + current branch.
func (s *Store) CreateGitRepository(ctx context.Context, projectID string, input CreateGitRepositoryInput, gitSvc gitwork.Service) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitRepository")
	return s.git.CreateGitRepository(ctx, projectID, input, gitSvc)
}

// DeleteGitRepository removes a repository when no running tasks reference it.
func (s *Store) DeleteGitRepository(ctx context.Context, projectID, repoID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteGitRepository")
	return s.git.DeleteGitRepository(ctx, projectID, repoID)
}

// ListGitWorktreesByRepo returns worktrees for a repository.
func (s *Store) ListGitWorktreesByRepo(ctx context.Context, repoID string) ([]gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListGitWorktreesByRepo")
	return s.git.ListGitWorktreesByRepo(ctx, repoID)
}

// ListGitWorktrees returns worktrees for a repository (project-scoped route compat).
func (s *Store) ListGitWorktrees(ctx context.Context, projectID, repoID string) ([]gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListGitWorktrees")
	return s.git.ListGitWorktrees(ctx, projectID, repoID)
}

// GetGitWorktree returns one worktree by ID.
func (s *Store) GetGitWorktree(ctx context.Context, projectID, worktreeID string) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitWorktree")
	return s.git.GetGitWorktree(ctx, projectID, worktreeID)
}

// GetGitWorktreeByID loads a worktree row by primary key.
func (s *Store) GetGitWorktreeByID(ctx context.Context, worktreeID string) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitWorktreeByID")
	return s.git.GetGitWorktreeByID(ctx, worktreeID)
}

// CreateGitWorktreeForRepo adds a worktree on disk and persists the row.
func (s *Store) CreateGitWorktreeForRepo(ctx context.Context, repoID string, input CreateGitWorktreeInput, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitWorktreeForRepo")
	return s.git.CreateGitWorktreeForRepo(ctx, repoID, input, gitSvc)
}

// CreateGitWorktree adds a worktree on disk and persists the row (project-scoped route compat).
func (s *Store) CreateGitWorktree(ctx context.Context, projectID, repoID string, input CreateGitWorktreeInput, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitWorktree")
	return s.git.CreateGitWorktree(ctx, projectID, repoID, input, gitSvc)
}

// RegisterExistingGitWorktree links a live checkout and optionally binds a branch.
func (s *Store) RegisterExistingGitWorktree(ctx context.Context, repoID, path, name string, bind BindBranchInput, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RegisterExistingGitWorktree")
	return s.git.RegisterExistingGitWorktree(ctx, repoID, path, name, bind, gitSvc)
}

// UnregisterGitWorktreeByID removes the Hamix row without deleting the checkout.
func (s *Store) UnregisterGitWorktreeByID(ctx context.Context, worktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UnregisterGitWorktreeByID")
	return s.git.UnregisterGitWorktreeByID(ctx, worktreeID)
}

// UnregisterGitWorktree removes the Hamix row without deleting the checkout.
func (s *Store) UnregisterGitWorktree(ctx context.Context, projectID, worktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UnregisterGitWorktree")
	return s.git.UnregisterGitWorktree(ctx, projectID, worktreeID)
}

// RemoveGitWorktreeFromDiskByID removes the checkout from disk and the Hamix row.
func (s *Store) RemoveGitWorktreeFromDiskByID(ctx context.Context, worktreeID string, force bool, gitSvc gitwork.Service) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RemoveGitWorktreeFromDiskByID")
	return s.git.RemoveGitWorktreeFromDiskByID(ctx, worktreeID, force, gitSvc)
}

// RemoveGitWorktreeFromDisk removes the checkout from disk and the Hamix row.
func (s *Store) RemoveGitWorktreeFromDisk(ctx context.Context, projectID, worktreeID string, force bool, gitSvc gitwork.Service) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RemoveGitWorktreeFromDisk")
	return s.git.RemoveGitWorktreeFromDisk(ctx, projectID, worktreeID, force, gitSvc)
}

// ListGitBranchesByRepo returns branches for a repository.
func (s *Store) ListGitBranchesByRepo(ctx context.Context, repoID string) ([]gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListGitBranchesByRepo")
	return s.git.ListGitBranchesByRepo(ctx, repoID)
}

// ListGitBranches returns branches for a repository (project-scoped route compat).
func (s *Store) ListGitBranches(ctx context.Context, projectID, repoID string) ([]gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListGitBranches")
	return s.git.ListGitBranches(ctx, projectID, repoID)
}

// GetGitBranch returns one branch by ID.
func (s *Store) GetGitBranch(ctx context.Context, projectID, branchID string) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitBranch")
	return s.git.GetGitBranch(ctx, projectID, branchID)
}

// GetGitBranchByID loads a branch row by primary key.
func (s *Store) GetGitBranchByID(ctx context.Context, branchID string) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitBranchByID")
	return s.git.GetGitBranchByID(ctx, branchID)
}

// CreateGitBranchForRepo creates a local branch via git.
func (s *Store) CreateGitBranchForRepo(ctx context.Context, repoID string, input CreateGitBranchInput, gitSvc gitwork.Service) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitBranchForRepo")
	return s.git.CreateGitBranchForRepo(ctx, repoID, input, gitSvc)
}

// CreateGitBranch creates a local branch via git (project-scoped route compat).
func (s *Store) CreateGitBranch(ctx context.Context, projectID, repoID string, input CreateGitBranchInput, gitSvc gitwork.Service) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitBranch")
	return s.git.CreateGitBranch(ctx, projectID, repoID, input, gitSvc)
}

// DeleteGitBranch deletes a branch row and optionally the git ref.
func (s *Store) DeleteGitBranch(ctx context.Context, projectID, branchID string, force bool, gitSvc gitwork.Service) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteGitBranch")
	return s.git.DeleteGitBranch(ctx, projectID, branchID, force, gitSvc)
}

// GuardBranchNotBoundToOtherWorktree rejects when branchID is already assigned to another worktree.
func (s *Store) GuardBranchNotBoundToOtherWorktree(ctx context.Context, branchID, exceptWorktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GuardBranchNotBoundToOtherWorktree")
	return s.git.GuardBranchNotBoundToOtherWorktree(ctx, branchID, exceptWorktreeID)
}

// RepoWorktreeInventory returns live git worktrees plus Hamix registration state.
func (s *Store) RepoWorktreeInventory(ctx context.Context, repo gitdomain.GitRepository, gitSvc gitwork.Service) ([]WorktreeInventoryRow, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RepoWorktreeInventory")
	return s.git.RepoWorktreeInventory(ctx, repo, gitSvc)
}

// RepoWorktreeCheckoutStatus returns live checkout git state for registered worktrees.
func (s *Store) RepoWorktreeCheckoutStatus(ctx context.Context, repo gitdomain.GitRepository, gitSvc gitwork.Service) ([]WorktreeCheckoutStatusRow, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RepoWorktreeCheckoutStatus")
	return s.git.RepoWorktreeCheckoutStatus(ctx, repo, gitSvc)
}

// ProbeGitWorktree describes whether a path is a linked, registerable worktree.
func (s *Store) ProbeGitWorktree(ctx context.Context, repoID, path string, gitSvc gitwork.Service) (GitWorktreeProbeResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ProbeGitWorktree")
	return s.git.ProbeGitWorktree(ctx, repoID, path, gitSvc)
}

// ReconcileGitRepository syncs repository/worktree paths with live git state.
func (s *Store) ReconcileGitRepository(ctx context.Context, projectID, repoID string, input ReconcileGitInput, gitSvc gitwork.Service) (ReconcileGitOutput, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ReconcileGitRepository")
	return s.git.ReconcileGitRepository(ctx, projectID, repoID, input, gitSvc)
}

// RelocateGitRepository updates the main repository path after a filesystem move.
func (s *Store) RelocateGitRepository(ctx context.Context, projectID, repoID, path string, gitSvc gitwork.Service) (ReconcileGitOutput, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RelocateGitRepository")
	return s.git.RelocateGitRepository(ctx, projectID, repoID, path, gitSvc)
}

// RelocateGitWorktree updates a registered worktree path after a filesystem move.
func (s *Store) RelocateGitWorktree(ctx context.Context, worktreeID, path string, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RelocateGitWorktree")
	return s.git.RelocateGitWorktree(ctx, worktreeID, path, gitSvc)
}

// ReconcileGitRepositoriesOnStartup runs best-effort reconcile for all registered repos.
func (s *Store) ReconcileGitRepositoriesOnStartup(ctx context.Context, gitSvc gitwork.Service) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ReconcileGitRepositoriesOnStartup")
	s.git.ReconcileGitRepositoriesOnStartup(ctx, gitSvc)
}
