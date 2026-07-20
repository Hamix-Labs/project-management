package store

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// TaskBranchName returns hamix/task-<first 8 hex of task id> (dashes stripped).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func TaskBranchName(taskID string) string {
	hex := strings.ReplaceAll(strings.TrimSpace(taskID), "-", "")
	if len(hex) > 8 {
		hex = hex[:8]
	}
	if hex == "" {
		hex = "00000000"
	}
	return "hamix/task-" + strings.ToLower(hex)
}

// AllocateTaskWorktree fetches origin, creates a linked worktree + branch for the
// task, and persists git_worktrees/git_branches rows (never is_main).
func (s *Store) AllocateTaskWorktree(ctx context.Context, repoID, taskID string, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.AllocateTaskWorktree")
	repoID = strings.TrimSpace(repoID)
	taskID = strings.TrimSpace(taskID)
	if repoID == "" {
		return gitdomain.GitWorktree{}, fmt.Errorf("%w: repository_id required", taskcoredomain.ErrInvalidInput)
	}
	if taskID == "" {
		return gitdomain.GitWorktree{}, fmt.Errorf("%w: task_id required", taskcoredomain.ErrInvalidInput)
	}
	repo, err := s.GetGitRepositoryByID(ctx, repoID)
	if err != nil {
		return gitdomain.GitWorktree{}, err
	}
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	opened, err := gitSvc.OpenRepository(ctx, repo.Path)
	if err != nil {
		return gitdomain.GitWorktree{}, fmt.Errorf("open repository: %w", err)
	}
	if err := gitSvc.Fetch(ctx, opened, "origin"); err != nil {
		return gitdomain.GitWorktree{}, fmt.Errorf("%w: fetch origin failed: %v", taskcoredomain.ErrInvalidInput, err)
	}
	defaultBranch := strings.TrimSpace(repo.DefaultBranch)
	if defaultBranch == "" {
		resolved, resolveErr := gitSvc.ResolveDefaultBranch(ctx, opened, "origin")
		if resolveErr != nil {
			return gitdomain.GitWorktree{}, fmt.Errorf("resolve default branch: %w", resolveErr)
		}
		defaultBranch = strings.TrimSpace(resolved)
	}
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	branch := TaskBranchName(taskID)
	if strings.EqualFold(branch, defaultBranch) {
		return gitdomain.GitWorktree{}, fmt.Errorf("%w: cannot allocate on default branch %q", taskcoredomain.ErrInvalidInput, defaultBranch)
	}
	startPoint := "origin/" + defaultBranch
	wtPath := gitinventory.ManagedWorktreePath(repo.Path, repo.ID, branch)
	if err := os.MkdirAll(filepath.FromSlash(filepath.Dir(wtPath)), 0o755); err != nil {
		return gitdomain.GitWorktree{}, fmt.Errorf("create managed worktree parent: %w", err)
	}
	return s.createGitWorktreeOnRepo(ctx, repo, CreateGitWorktreeInput{
		Path:         filepath.FromSlash(wtPath),
		Name:         branch,
		Branch:       branch,
		CreateBranch: true,
		StartPoint:   startPoint,
	}, gitSvc)
}
