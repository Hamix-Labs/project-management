package store

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// WorktreeStaleAfter is how long after the latest terminal task a managed
// worktree is considered stale (no settings UI in this epic).
const WorktreeStaleAfter = 24 * time.Hour

// SyncGitRepository fetches origin and refreshes Hamix metadata without
// discovering/registering operator worktrees.
func (s *Store) SyncGitRepository(ctx context.Context, repoID string) (contract.ReconcileGitOutput, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.SyncGitRepository")
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		return contract.ReconcileGitOutput{}, fmt.Errorf("%w: repository_id required", taskcoredomain.ErrInvalidInput)
	}
	repo, err := s.GetGitRepositoryByID(ctx, repoID)
	if err != nil {
		return contract.ReconcileGitOutput{}, err
	}
	opened, err := s.gitSvc().OpenRepository(ctx, repo.Path)
	if err != nil {
		return contract.ReconcileGitOutput{}, fmt.Errorf("open repository: %w", err)
	}
	if err := s.gitSvc().Fetch(ctx, opened, "origin"); err != nil {
		return contract.ReconcileGitOutput{}, fmt.Errorf("%w: fetch origin failed: %v", taskcoredomain.ErrInvalidInput, err)
	}
	return s.ReconcileGitRepository(ctx, "", repoID, contract.ReconcileGitInput{
		AllowRemove:           true,
		AllowCheckoutDiscover: false,
		AllowDiscover:         false,
	})
}

// WorktreeStaleMap reports which worktree IDs are stale for the given repo.
func (s *Store) WorktreeStaleMap(ctx context.Context, repoID string, now time.Time) (map[string]bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.WorktreeStaleMap")
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		return nil, fmt.Errorf("%w: repository_id required", taskcoredomain.ErrInvalidInput)
	}
	wts, err := s.ListGitWorktreesByRepo(ctx, repoID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(wts))
	cutoff := now.UTC().Add(-WorktreeStaleAfter)
	terminal := []string{string(taskcoredomain.StatusDone), string(taskcoredomain.StatusFailed)}
	nonTerminal := []string{
		string(taskcoredomain.StatusReady),
		string(taskcoredomain.StatusRunning),
		string(taskcoredomain.StatusBlocked),
		string(taskcoredomain.StatusReview),
		string(taskcoredomain.StatusOnHold),
	}
	for _, wt := range wts {
		if wt.IsMain {
			out[wt.ID] = false
			continue
		}
		var active int64
		if err := s.db.WithContext(ctx).Table("tasks").
			Where("worktree_id = ? AND status IN ?", wt.ID, nonTerminal).
			Count(&active).Error; err != nil {
			return nil, fmt.Errorf("count active tasks: %w", err)
		}
		if active > 0 {
			out[wt.ID] = false
			continue
		}
		var latest *time.Time
		row := s.db.WithContext(ctx).Table("tasks").
			Select("MAX(updated_at)").
			Where("worktree_id = ? AND status IN ?", wt.ID, terminal).
			Row()
		if err := row.Scan(&latest); err != nil {
			return nil, fmt.Errorf("latest terminal task: %w", err)
		}
		if latest == nil {
			out[wt.ID] = false
			continue
		}
		out[wt.ID] = latest.UTC().Before(cutoff)
	}
	return out, nil
}
