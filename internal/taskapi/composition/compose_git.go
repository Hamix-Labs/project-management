package composition

import (
	"context"
	"log/slog"
	"time"

	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

// FindWorktreeInInventory delegates to the git inventory store helper.
var FindWorktreeInInventory = gitinventorystore.FindWorktreeInInventory

// ListAllGitRepositories returns every registered repository ordered by created_at.
func (a *API) ListAllGitRepositories(ctx context.Context) ([]gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListAllGitRepositories")
	return a.git.ListAllGitRepositories(ctx)
}

// ListAllGitRepositoriesWithSummary returns repositories with list-page metadata.
func (a *API) ListAllGitRepositoriesWithSummary(ctx context.Context) ([]gitinventorystore.GitRepositoryListSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListAllGitRepositoriesWithSummary")
	return a.git.ListAllGitRepositoriesWithSummary(ctx)
}

// CreateGlobalGitRepository registers a repository without project scoping.
func (a *API) CreateGlobalGitRepository(ctx context.Context, input gitinventorystore.CreateGitRepositoryInput, gitSvc gitwork.Service) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGlobalGitRepository")
	return a.git.CreateGlobalGitRepository(ctx, input, gitSvc)
}

// DeleteGlobalGitRepository removes a repository by ID.
func (a *API) DeleteGlobalGitRepository(ctx context.Context, repoID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteGlobalGitRepository")
	return a.git.DeleteGlobalGitRepository(ctx, repoID)
}

// ListGitRepositories returns all registered repositories ordered by created_at.
func (a *API) ListGitRepositories(ctx context.Context, projectID string) ([]gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListGitRepositories")
	return a.git.ListGitRepositories(ctx, projectID)
}

// CountGitRepositories returns the total number of registered git repositories.
func (a *API) CountGitRepositories(ctx context.Context) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CountGitRepositories")
	return a.git.CountGitRepositories(ctx)
}

// GetGitRepository returns one repository by ID.
func (a *API) GetGitRepository(ctx context.Context, projectID, repoID string) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitRepository")
	return a.git.GetGitRepository(ctx, projectID, repoID)
}

// GetGitRepositoryByID loads a repository row by primary key.
func (a *API) GetGitRepositoryByID(ctx context.Context, repoID string) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitRepositoryByID")
	return a.git.GetGitRepositoryByID(ctx, repoID)
}

// CreateGitRepository validates path with git, then inserts repository + main worktree + current branch.
func (a *API) CreateGitRepository(ctx context.Context, projectID string, input gitinventorystore.CreateGitRepositoryInput, gitSvc gitwork.Service) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitRepository")
	return a.git.CreateGitRepository(ctx, projectID, input, gitSvc)
}

// DeleteGitRepository removes a repository when no running tasks reference it.
func (a *API) DeleteGitRepository(ctx context.Context, projectID, repoID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteGitRepository")
	return a.git.DeleteGitRepository(ctx, projectID, repoID)
}

// ListGitWorktreesByRepo returns worktrees for a repository.
func (a *API) ListGitWorktreesByRepo(ctx context.Context, repoID string) ([]gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListGitWorktreesByRepo")
	return a.git.ListGitWorktreesByRepo(ctx, repoID)
}

// ListGitWorktrees returns worktrees for a repository (project-scoped route compat).
func (a *API) ListGitWorktrees(ctx context.Context, projectID, repoID string) ([]gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListGitWorktrees")
	return a.git.ListGitWorktrees(ctx, projectID, repoID)
}

// GetGitWorktree returns one worktree by ID.
func (a *API) GetGitWorktree(ctx context.Context, projectID, worktreeID string) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitWorktree")
	return a.git.GetGitWorktree(ctx, projectID, worktreeID)
}

// GetGitWorktreeByID loads a worktree row by primary key.
func (a *API) GetGitWorktreeByID(ctx context.Context, worktreeID string) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitWorktreeByID")
	return a.git.GetGitWorktreeByID(ctx, worktreeID)
}

// CreateGitWorktreeForRepo adds a worktree on disk and persists the row.
func (a *API) CreateGitWorktreeForRepo(ctx context.Context, repoID string, input gitinventorystore.CreateGitWorktreeInput, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitWorktreeForRepo")
	return a.git.CreateGitWorktreeForRepo(ctx, repoID, input, gitSvc)
}

// AllocateTaskWorktree fetches origin and creates a Hamix-managed worktree for a task.
func (a *API) AllocateTaskWorktree(ctx context.Context, repoID, taskID string, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AllocateTaskWorktree")
	return a.git.AllocateTaskWorktree(ctx, repoID, taskID, gitSvc)
}

// SyncGitRepository fetches origin and refreshes metadata without discover.
func (a *API) SyncGitRepository(ctx context.Context, repoID string, gitSvc gitwork.Service) (gitinventorystore.ReconcileGitOutput, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SyncGitRepository")
	return a.git.SyncGitRepository(ctx, repoID, gitSvc)
}

// WorktreeStaleMap reports stale managed worktrees for a repository.
func (a *API) WorktreeStaleMap(ctx context.Context, repoID string, now time.Time) (map[string]bool, error) {
	return a.git.WorktreeStaleMap(ctx, repoID, now)
}

// CreateGitWorktree adds a worktree on disk and persists the row (project-scoped route compat).
func (a *API) CreateGitWorktree(ctx context.Context, projectID, repoID string, input gitinventorystore.CreateGitWorktreeInput, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitWorktree")
	return a.git.CreateGitWorktree(ctx, projectID, repoID, input, gitSvc)
}

// RegisterExistingGitWorktree links a live checkout and optionally binds a branch.
func (a *API) RegisterExistingGitWorktree(ctx context.Context, repoID, path, name string, bind gitinventorystore.BindBranchInput, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RegisterExistingGitWorktree")
	return a.git.RegisterExistingGitWorktree(ctx, repoID, path, name, bind, gitSvc)
}

// UnregisterGitWorktreeByID removes the Hamix row without deleting the checkout.
func (a *API) UnregisterGitWorktreeByID(ctx context.Context, worktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UnregisterGitWorktreeByID")
	return a.git.UnregisterGitWorktreeByID(ctx, worktreeID)
}

// UnregisterGitWorktree removes the Hamix row without deleting the checkout.
func (a *API) UnregisterGitWorktree(ctx context.Context, projectID, worktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.UnregisterGitWorktree")
	return a.git.UnregisterGitWorktree(ctx, projectID, worktreeID)
}

// RemoveGitWorktreeFromDiskByID removes the checkout from disk and the Hamix row.
func (a *API) RemoveGitWorktreeFromDiskByID(ctx context.Context, worktreeID string, force bool, gitSvc gitwork.Service) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RemoveGitWorktreeFromDiskByID")
	return a.git.RemoveGitWorktreeFromDiskByID(ctx, worktreeID, force, gitSvc)
}

// RemoveGitWorktreeFromDisk removes the checkout from disk and the Hamix row.
func (a *API) RemoveGitWorktreeFromDisk(ctx context.Context, projectID, worktreeID string, force bool, gitSvc gitwork.Service) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RemoveGitWorktreeFromDisk")
	return a.git.RemoveGitWorktreeFromDisk(ctx, projectID, worktreeID, force, gitSvc)
}

// ListGitBranchesByRepo returns branches for a repository.
func (a *API) ListGitBranchesByRepo(ctx context.Context, repoID string) ([]gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListGitBranchesByRepo")
	return a.git.ListGitBranchesByRepo(ctx, repoID)
}

// ListGitBranches returns branches for a repository (project-scoped route compat).
func (a *API) ListGitBranches(ctx context.Context, projectID, repoID string) ([]gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ListGitBranches")
	return a.git.ListGitBranches(ctx, projectID, repoID)
}

// GetGitBranch returns one branch by ID.
func (a *API) GetGitBranch(ctx context.Context, projectID, branchID string) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitBranch")
	return a.git.GetGitBranch(ctx, projectID, branchID)
}

// GetGitBranchByID loads a branch row by primary key.
func (a *API) GetGitBranchByID(ctx context.Context, branchID string) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GetGitBranchByID")
	return a.git.GetGitBranchByID(ctx, branchID)
}

// CreateGitBranchForRepo creates a local branch via git.
func (a *API) CreateGitBranchForRepo(ctx context.Context, repoID string, input gitinventorystore.CreateGitBranchInput, gitSvc gitwork.Service) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitBranchForRepo")
	return a.git.CreateGitBranchForRepo(ctx, repoID, input, gitSvc)
}

// CreateGitBranch creates a local branch via git (project-scoped route compat).
func (a *API) CreateGitBranch(ctx context.Context, projectID, repoID string, input gitinventorystore.CreateGitBranchInput, gitSvc gitwork.Service) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitBranch")
	return a.git.CreateGitBranch(ctx, projectID, repoID, input, gitSvc)
}

// DeleteGitBranch deletes a branch row and optionally the git ref.
func (a *API) DeleteGitBranch(ctx context.Context, projectID, branchID string, force bool, gitSvc gitwork.Service) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteGitBranch")
	return a.git.DeleteGitBranch(ctx, projectID, branchID, force, gitSvc)
}

// GuardBranchNotBoundToOtherWorktree rejects when branchID is already assigned to another worktree.
func (a *API) GuardBranchNotBoundToOtherWorktree(ctx context.Context, branchID, exceptWorktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GuardBranchNotBoundToOtherWorktree")
	return a.git.GuardBranchNotBoundToOtherWorktree(ctx, branchID, exceptWorktreeID)
}

// RepoWorktreeInventory returns live git worktrees plus Hamix registration state.
func (a *API) RepoWorktreeInventory(ctx context.Context, repo gitdomain.GitRepository, gitSvc gitwork.Service) ([]gitinventorystore.WorktreeInventoryRow, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RepoWorktreeInventory")
	return a.git.RepoWorktreeInventory(ctx, repo, gitSvc)
}

// RepoWorktreeCheckoutStatus returns live checkout git state for registered worktrees.
func (a *API) RepoWorktreeCheckoutStatus(ctx context.Context, repo gitdomain.GitRepository, gitSvc gitwork.Service) ([]gitinventorystore.WorktreeCheckoutStatusRow, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RepoWorktreeCheckoutStatus")
	return a.git.RepoWorktreeCheckoutStatus(ctx, repo, gitSvc)
}

// ProbeGitWorktree describes whether a path is a linked, registerable worktree.
func (a *API) ProbeGitWorktree(ctx context.Context, repoID, path string, gitSvc gitwork.Service) (gitinventorystore.GitWorktreeProbeResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ProbeGitWorktree")
	return a.git.ProbeGitWorktree(ctx, repoID, path, gitSvc)
}

// ReconcileGitRepository syncs repository/worktree paths with live git state.
func (a *API) ReconcileGitRepository(ctx context.Context, projectID, repoID string, input gitinventorystore.ReconcileGitInput, gitSvc gitwork.Service) (gitinventorystore.ReconcileGitOutput, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ReconcileGitRepository")
	return a.git.ReconcileGitRepository(ctx, projectID, repoID, input, gitSvc)
}

// RelocateGitRepository updates the main repository path after a filesystem move.
func (a *API) RelocateGitRepository(ctx context.Context, projectID, repoID, path string, gitSvc gitwork.Service) (gitinventorystore.ReconcileGitOutput, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RelocateGitRepository")
	return a.git.RelocateGitRepository(ctx, projectID, repoID, path, gitSvc)
}

// RelocateGitWorktree updates a registered worktree path after a filesystem move.
func (a *API) RelocateGitWorktree(ctx context.Context, worktreeID, path string, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RelocateGitWorktree")
	return a.git.RelocateGitWorktree(ctx, worktreeID, path, gitSvc)
}

// ReconcileGitRepositoriesOnStartup runs best-effort reconcile for all registered repos.
func (a *API) ReconcileGitRepositoriesOnStartup(ctx context.Context, gitSvc gitwork.Service) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ReconcileGitRepositoriesOnStartup")
	a.git.ReconcileGitRepositoriesOnStartup(ctx, gitSvc)
}
