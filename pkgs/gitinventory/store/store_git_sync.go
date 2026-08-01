package store

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
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

// WorktreeStaleMap reports which of the given worktree IDs are stale.
// Query count is O(1) in len(worktrees): two set queries (active tasks,
// latest terminal cycle end), then in-memory classification.
func (s *Store) WorktreeStaleMap(ctx context.Context, worktrees []gitdomain.GitWorktree, now time.Time) (map[string]bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.WorktreeStaleMap")
	out := make(map[string]bool, len(worktrees))
	if len(worktrees) == 0 {
		return out, nil
	}

	cutoff := now.UTC().Add(-WorktreeStaleAfter)
	candidateIDs := make([]string, 0, len(worktrees))
	for _, wt := range worktrees {
		out[wt.ID] = false
		if wt.IsMain {
			continue
		}
		candidateIDs = append(candidateIDs, wt.ID)
	}
	if len(candidateIDs) == 0 {
		return out, nil
	}

	terminalCycles := []string{
		string(cyclesdomain.CycleStatusSucceeded),
		string(cyclesdomain.CycleStatusFailed),
		string(cyclesdomain.CycleStatusAborted),
	}
	nonTerminal := []string{
		string(taskcoredomain.StatusReady),
		string(taskcoredomain.StatusRunning),
		string(taskcoredomain.StatusBlocked),
		string(taskcoredomain.StatusReview),
		string(taskcoredomain.StatusPrReady),
		string(taskcoredomain.StatusOnHold),
	}

	var activeIDs []string
	if err := s.db.WithContext(ctx).Table("tasks").
		Where("worktree_id IN ? AND status IN ?", candidateIDs, nonTerminal).
		Distinct("worktree_id").
		Pluck("worktree_id", &activeIDs).Error; err != nil {
		return nil, fmt.Errorf("count active tasks: %w", err)
	}
	active := make(map[string]struct{}, len(activeIDs))
	for _, id := range activeIDs {
		active[id] = struct{}{}
	}

	type latestRow struct {
		WorktreeID string `gorm:"column:worktree_id"`
		Latest     string `gorm:"column:latest"`
	}
	var latestRows []latestRow
	if err := s.db.WithContext(ctx).
		Table("task_cycles AS c").
		Select("t.worktree_id AS worktree_id, MAX(c.ended_at) AS latest").
		Joins("INNER JOIN tasks AS t ON t.id = c.task_id").
		Where("t.worktree_id IN ? AND c.status IN ? AND c.ended_at IS NOT NULL", candidateIDs, terminalCycles).
		Group("t.worktree_id").
		Scan(&latestRows).Error; err != nil {
		return nil, fmt.Errorf("latest terminal cycle: %w", err)
	}
	latestByID := make(map[string]time.Time, len(latestRows))
	for _, row := range latestRows {
		if strings.TrimSpace(row.Latest) == "" {
			continue
		}
		latest, err := parseSQLiteDateTime(row.Latest)
		if err != nil {
			return nil, fmt.Errorf("latest terminal cycle: parse %q: %w", row.Latest, err)
		}
		latestByID[row.WorktreeID] = latest
	}

	for _, id := range candidateIDs {
		if _, ok := active[id]; ok {
			out[id] = false
			continue
		}
		latest, ok := latestByID[id]
		if !ok {
			out[id] = false
			continue
		}
		out[id] = latest.UTC().Before(cutoff)
	}
	return out, nil
}

//funclogmeasure:skip category=hot-path reason="Pure datetime parse helper; callers emit operation traces."
func parseSQLiteDateTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999+00:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	var firstErr error
	for _, layout := range layouts {
		t, err := time.Parse(layout, raw)
		if err == nil {
			return t, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return time.Time{}, firstErr
}
