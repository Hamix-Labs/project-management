package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"fmt"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"log/slog"
	"os"
	"strings"
)

// TaskGitContext is the resolved filesystem path and branch name for a task binding.
type TaskGitContext struct {
	WorktreeID   string
	BranchID     string
	WorktreePath string
	BranchName   string
}

// ValidateTaskWorktreeBinding checks worktree_id exists and, when projectID is
// set, that project.repository_id matches the worktree's repo.
func (s *Store) ValidateTaskWorktreeBinding(ctx context.Context, projectID *string, worktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ValidateTaskWorktreeBinding")
	worktreeID = strings.TrimSpace(worktreeID)
	if worktreeID == "" {
		return fmt.Errorf("%w: worktree_id required", taskcoredomain.ErrInvalidInput)
	}
	wt, err := s.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(wt.BranchID) == "" {
		return fmt.Errorf("%w: worktree has no branch assigned", taskcoredomain.ErrInvalidInput)
	}
	if _, err := s.GetGitBranchByID(ctx, wt.BranchID); err != nil {
		return err
	}
	if projectID == nil {
		return nil
	}
	pid := strings.TrimSpace(*projectID)
	if pid == "" {
		return nil
	}
	proj, err := s.GetProject(ctx, pid)
	if err != nil {
		return err
	}
	if proj.RepositoryID == nil || *proj.RepositoryID != wt.RepositoryID {
		return gitdomain.NewGitErr(gitdomain.GitCodeProjectRepoMismatch, "project is not tied to this repository")
	}
	return nil
}

// ResolveTaskGitContext loads worktree path and branch name via worktree_id.
func (s *Store) ResolveTaskGitContext(ctx context.Context, worktreeID string) (TaskGitContext, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.ResolveTaskGitContext")
	if err := s.ValidateTaskWorktreeBinding(ctx, nil, worktreeID); err != nil {
		return TaskGitContext{}, err
	}
	wt, err := s.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil {
		return TaskGitContext{}, err
	}
	br, err := s.GetGitBranchByID(ctx, wt.BranchID)
	if err != nil {
		return TaskGitContext{}, err
	}
	return TaskGitContext{
		WorktreeID:   wt.ID,
		BranchID:     br.ID,
		WorktreePath: wt.Path,
		BranchName:   br.Name,
	}, nil
}

// AgentWorkerGitIdle reports whether the worker should stay idle for git registration reasons.
func (s *Store) AgentWorkerGitIdle(ctx context.Context) (idle bool, reason string, err error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.AgentWorkerGitIdle")
	repoCount, err := s.CountGitRepositories(ctx)
	if err != nil {
		return false, "", err
	}
	if repoCount == 0 {
		return true, "no_repository_registered", nil
	}
	repos, err := s.ListAllGitRepositories(ctx)
	if err != nil {
		return false, "", err
	}
	var worktrees []gitdomain.GitWorktree
	for _, repo := range repos {
		wts, wtErr := s.ListGitWorktreesByRepo(ctx, repo.ID)
		if wtErr != nil {
			return false, "", wtErr
		}
		worktrees = append(worktrees, wts...)
	}
	if len(worktrees) == 0 {
		return true, "all_worktrees_invalid", nil
	}
	anyValid := false
	for _, wt := range worktrees {
		st, statErr := os.Stat(wt.Path)
		if statErr == nil && st.IsDir() {
			anyValid = true
			break
		}
	}
	if !anyValid {
		return true, "all_worktrees_invalid", nil
	}
	return false, "", nil
}
