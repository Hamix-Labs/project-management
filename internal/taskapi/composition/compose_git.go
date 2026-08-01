package composition

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
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

// CreateGlobalGitRepository registers a repository without project scoping and
// seeds the system default project (cross-BC write owned by composition).
func (a *API) CreateGlobalGitRepository(ctx context.Context, input gitinventorystore.CreateGitRepositoryInput) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGlobalGitRepository")
	repo, err := a.git.CreateGlobalGitRepository(ctx, input)
	if err != nil {
		return gitdomain.GitRepository{}, err
	}
	if _, err := a.projects.CreateDefaultProjectForRepo(ctx, repo.ID); err != nil {
		return gitdomain.GitRepository{}, fmt.Errorf("seed default project: %w", err)
	}
	return repo, nil
}

// DeleteGlobalGitRepository removes projects for the repo, then the repository.
func (a *API) DeleteGlobalGitRepository(ctx context.Context, repoID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteGlobalGitRepository")
	if err := a.projects.DeleteProjectsForRepository(ctx, repoID); err != nil {
		return err
	}
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

// CreateGitRepository validates path with git, then inserts repository + main worktree + current branch
// and seeds the system default project (cross-BC write owned by composition).
func (a *API) CreateGitRepository(ctx context.Context, projectID string, input gitinventorystore.CreateGitRepositoryInput) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitRepository")
	repo, err := a.git.CreateGitRepository(ctx, projectID, input)
	if err != nil {
		return gitdomain.GitRepository{}, err
	}
	if _, err := a.projects.CreateDefaultProjectForRepo(ctx, repo.ID); err != nil {
		return gitdomain.GitRepository{}, fmt.Errorf("seed default project: %w", err)
	}
	return repo, nil
}

// DeleteGitRepository removes projects for the repo, then the repository.
func (a *API) DeleteGitRepository(ctx context.Context, projectID, repoID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteGitRepository")
	if err := a.projects.DeleteProjectsForRepository(ctx, repoID); err != nil {
		return err
	}
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
func (a *API) CreateGitWorktreeForRepo(ctx context.Context, repoID string, input gitinventorystore.CreateGitWorktreeInput) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitWorktreeForRepo")
	return a.git.CreateGitWorktreeForRepo(ctx, repoID, input)
}

// AllocateTaskWorktree fetches origin and creates a Hamix-managed worktree for a task.
func (a *API) AllocateTaskWorktree(ctx context.Context, repoID, taskID string) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AllocateTaskWorktree")
	return a.git.AllocateTaskWorktree(ctx, repoID, taskID)
}

// SyncGitRepository fetches origin and refreshes metadata without discover.
func (a *API) SyncGitRepository(ctx context.Context, repoID string) (gitinventorystore.ReconcileGitOutput, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SyncGitRepository")
	return a.git.SyncGitRepository(ctx, repoID)
}

// WorktreeStaleMap reports stale flags for the given worktrees.
//
//funclogmeasure:skip category=delegate-already-logs reason="Thin store facade; gitinventory.store.WorktreeStaleMap emits operation trace."
func (a *API) WorktreeStaleMap(ctx context.Context, worktrees []gitdomain.GitWorktree, now time.Time) (map[string]bool, error) {
	return a.git.WorktreeStaleMap(ctx, worktrees, now)
}

// CreateGitWorktree adds a worktree on disk and persists the row (project-scoped route compat).
func (a *API) CreateGitWorktree(ctx context.Context, projectID, repoID string, input gitinventorystore.CreateGitWorktreeInput) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitWorktree")
	return a.git.CreateGitWorktree(ctx, projectID, repoID, input)
}

// RegisterExistingGitWorktree links a live checkout and optionally binds a branch.
func (a *API) RegisterExistingGitWorktree(ctx context.Context, repoID, path, name string, bind gitinventorystore.BindBranchInput) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RegisterExistingGitWorktree")
	return a.git.RegisterExistingGitWorktree(ctx, repoID, path, name, bind)
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
func (a *API) RemoveGitWorktreeFromDiskByID(ctx context.Context, worktreeID string, force bool) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RemoveGitWorktreeFromDiskByID")
	return a.git.RemoveGitWorktreeFromDiskByID(ctx, worktreeID, force)
}

// RemoveGitWorktreeFromDisk removes the checkout from disk and the Hamix row.
func (a *API) RemoveGitWorktreeFromDisk(ctx context.Context, projectID, worktreeID string, force bool) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RemoveGitWorktreeFromDisk")
	return a.git.RemoveGitWorktreeFromDisk(ctx, projectID, worktreeID, force)
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
func (a *API) CreateGitBranchForRepo(ctx context.Context, repoID string, input gitinventorystore.CreateGitBranchInput) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitBranchForRepo")
	return a.git.CreateGitBranchForRepo(ctx, repoID, input)
}

// CreateGitBranch creates a local branch via git (project-scoped route compat).
func (a *API) CreateGitBranch(ctx context.Context, projectID, repoID string, input gitinventorystore.CreateGitBranchInput) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.CreateGitBranch")
	return a.git.CreateGitBranch(ctx, projectID, repoID, input)
}

// DeleteGitBranch deletes a branch row and optionally the git ref.
func (a *API) DeleteGitBranch(ctx context.Context, projectID, branchID string, force bool) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteGitBranch")
	return a.git.DeleteGitBranch(ctx, projectID, branchID, force)
}

// DeleteGitBranchByID deletes a branch by primary key (global route).
func (a *API) DeleteGitBranchByID(ctx context.Context, branchID string, force bool) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.DeleteGitBranchByID")
	return a.git.DeleteGitBranchByID(ctx, branchID, force)
}

// GuardBranchNotBoundToOtherWorktree rejects when branchID is already assigned to another worktree.
func (a *API) GuardBranchNotBoundToOtherWorktree(ctx context.Context, branchID, exceptWorktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.GuardBranchNotBoundToOtherWorktree")
	return a.git.GuardBranchNotBoundToOtherWorktree(ctx, branchID, exceptWorktreeID)
}

// RepoWorktreeInventory returns live git worktrees plus Hamix registration state.
func (a *API) RepoWorktreeInventory(ctx context.Context, repo gitdomain.GitRepository) ([]gitinventorystore.WorktreeInventoryRow, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RepoWorktreeInventory")
	return a.git.RepoWorktreeInventory(ctx, repo)
}

// RepoWorktreeCheckoutStatus returns live checkout git state for registered worktrees.
func (a *API) RepoWorktreeCheckoutStatus(ctx context.Context, repo gitdomain.GitRepository) ([]gitinventorystore.WorktreeCheckoutStatusRow, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RepoWorktreeCheckoutStatus")
	return a.git.RepoWorktreeCheckoutStatus(ctx, repo)
}

// ProbeGitWorktree describes whether a path is a linked, registerable worktree.
func (a *API) ProbeGitWorktree(ctx context.Context, repoID, path string) (gitinventorystore.GitWorktreeProbeResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ProbeGitWorktree")
	return a.git.ProbeGitWorktree(ctx, repoID, path)
}

// ReconcileGitRepository syncs repository/worktree paths with live git state.
func (a *API) ReconcileGitRepository(ctx context.Context, projectID, repoID string, input gitinventorystore.ReconcileGitInput) (gitinventorystore.ReconcileGitOutput, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ReconcileGitRepository")
	return a.git.ReconcileGitRepository(ctx, projectID, repoID, input)
}

// RelocateGitRepository updates the main repository path after a filesystem move.
func (a *API) RelocateGitRepository(ctx context.Context, projectID, repoID, path string) (gitinventorystore.ReconcileGitOutput, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RelocateGitRepository")
	return a.git.RelocateGitRepository(ctx, projectID, repoID, path)
}

// RelocateGitWorktree updates a registered worktree path after a filesystem move.
func (a *API) RelocateGitWorktree(ctx context.Context, worktreeID, path string) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.RelocateGitWorktree")
	return a.git.RelocateGitWorktree(ctx, worktreeID, path)
}

// ReconcileGitRepositoriesOnStartup runs best-effort reconcile for all registered repos.
func (a *API) ReconcileGitRepositoriesOnStartup(ctx context.Context) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ReconcileGitRepositoriesOnStartup")
	a.git.ReconcileGitRepositoriesOnStartup(ctx)
}
