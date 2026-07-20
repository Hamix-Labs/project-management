package storehooks

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	projectsstore "github.com/AlexsanderHamir/Hamix/pkgs/projects/store"
	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// GitDeps is the BC store surface required for task git-context resolution.
type GitDeps struct {
	Git      *gitinventorystore.Store
	Projects *projectsstore.Store
}

// ValidateTaskWorktreeBinding checks worktree_id exists and, when projectID is
// set, that project.repository_id matches the worktree's repo.
func ValidateTaskWorktreeBinding(ctx context.Context, d GitDeps, projectID *string, worktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.storehooks.ValidateTaskWorktreeBinding")
	worktreeID = strings.TrimSpace(worktreeID)
	if worktreeID == "" {
		return fmt.Errorf("%w: worktree_id required", taskcoredomain.ErrInvalidInput)
	}
	wt, err := d.Git.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil {
		return err
	}
	if wt.IsMain {
		return fmt.Errorf("%w: cannot bind task to main worktree", taskcoredomain.ErrInvalidInput)
	}
	if strings.TrimSpace(wt.BranchID) == "" {
		return fmt.Errorf("%w: worktree has no branch assigned", taskcoredomain.ErrInvalidInput)
	}
	br, err := d.Git.GetGitBranchByID(ctx, wt.BranchID)
	if err != nil {
		return err
	}
	repo, err := d.Git.GetGitRepositoryByID(ctx, wt.RepositoryID)
	if err != nil {
		return err
	}
	defaultBranch := strings.TrimSpace(repo.DefaultBranch)
	if defaultBranch != "" && strings.EqualFold(strings.TrimSpace(br.Name), defaultBranch) {
		return fmt.Errorf("%w: cannot bind task to default branch %q", taskcoredomain.ErrInvalidInput, defaultBranch)
	}
	if projectID == nil {
		return nil
	}
	pid := strings.TrimSpace(*projectID)
	if pid == "" {
		return nil
	}
	proj, err := d.Projects.GetProject(ctx, pid)
	if err != nil {
		return err
	}
	if proj.RepositoryID == nil || *proj.RepositoryID != wt.RepositoryID {
		return gitdomain.NewGitErr(gitdomain.GitCodeProjectRepoMismatch, "project is not tied to this repository")
	}
	return nil
}

// ResolveTaskGitContext loads worktree path and branch name via worktree_id.
func ResolveTaskGitContext(ctx context.Context, d GitDeps, worktreeID string) (taskcorecontract.TaskGitContext, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.storehooks.ResolveTaskGitContext")
	if err := ValidateTaskWorktreeBinding(ctx, d, nil, worktreeID); err != nil {
		return taskcorecontract.TaskGitContext{}, err
	}
	wt, err := d.Git.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil {
		return taskcorecontract.TaskGitContext{}, err
	}
	br, err := d.Git.GetGitBranchByID(ctx, wt.BranchID)
	if err != nil {
		return taskcorecontract.TaskGitContext{}, err
	}
	return taskcorecontract.TaskGitContext{
		WorktreeID:   wt.ID,
		BranchID:     br.ID,
		WorktreePath: wt.Path,
		BranchName:   br.Name,
	}, nil
}

// AgentWorkerGitIdle reports whether the worker should stay idle for git registration reasons.
func AgentWorkerGitIdle(ctx context.Context, git *gitinventorystore.Store) (idle bool, reason string, err error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskapi.storehooks.AgentWorkerGitIdle")
	repoCount, err := git.CountGitRepositories(ctx)
	if err != nil {
		return false, "", err
	}
	if repoCount == 0 {
		return true, "no_repository_registered", nil
	}
	repos, err := git.ListAllGitRepositories(ctx)
	if err != nil {
		return false, "", err
	}
	var worktrees []gitdomain.GitWorktree
	for _, repo := range repos {
		wts, wtErr := git.ListGitWorktreesByRepo(ctx, repo.ID)
		if wtErr != nil {
			return false, "", wtErr
		}
		worktrees = append(worktrees, wts...)
	}
	if len(worktrees) == 0 {
		return true, "all_worktrees_invalid", nil
	}
	for _, wt := range worktrees {
		st, statErr := os.Stat(wt.Path)
		if statErr == nil && st.IsDir() {
			return false, "", nil
		}
	}
	return true, "all_worktrees_invalid", nil
}
