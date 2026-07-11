package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"fmt"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// gitWorktreeIsFullyRegistered reports whether Hamix has a branch-bound worktree row.
// Reconcile discover may insert path-only rows (empty branch_id); those are not
// operator-registered and must remain selectable in live inventory.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by RepoWorktreeInventory."
func gitWorktreeIsFullyRegistered(wt gitdomain.GitWorktree) bool {
	return strings.TrimSpace(wt.BranchID) != ""
}

// findGitWorktreeByRepoPath returns a registered row for repo+path, if any.
func (s *Store) findGitWorktreeByRepoPath(
	ctx context.Context,
	repoID, path string,
) (gitdomain.GitWorktree, bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.findGitWorktreeByRepoPath")
	rows, err := s.ListGitWorktreesByRepo(ctx, repoID)
	if err != nil {
		return gitdomain.GitWorktree{}, false, err
	}
	want := worktreePathKey(path)
	for _, row := range rows {
		if worktreePathKey(row.Path) == want {
			return row, true, nil
		}
	}
	return gitdomain.GitWorktree{}, false, nil
}

// WorktreeInventoryRow is a live git worktree plus Hamix registration state.
type WorktreeInventoryRow = contract.WorktreeInventoryRow

// GitWorktreeProbeResult describes whether a path is a linked, registerable worktree.
type GitWorktreeProbeResult = contract.GitWorktreeProbeResult

// RepoWorktreeInventory lists live git worktrees for a repository and marks registered paths.
func (s *Store) RepoWorktreeInventory(
	ctx context.Context,
	repo gitdomain.GitRepository,
	gitSvc gitwork.Service,
) ([]WorktreeInventoryRow, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.RepoWorktreeInventory")
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	registered, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		return nil, err
	}
	registeredPaths := make(map[string]struct{}, len(registered))
	for _, wt := range registered {
		if !gitWorktreeIsFullyRegistered(wt) {
			continue
		}
		registeredPaths[worktreePathKey(wt.Path)] = struct{}{}
	}
	opened, err := gitSvc.OpenRepository(ctx, repo.Path)
	if err != nil {
		return nil, fmt.Errorf("open repository: %w", err)
	}
	live, err := gitSvc.ListWorktrees(ctx, opened)
	if err != nil {
		return nil, fmt.Errorf("list worktrees: %w", err)
	}
	out := make([]WorktreeInventoryRow, 0, len(live))
	for _, wt := range live {
		_, isRegistered := registeredPaths[worktreePathKey(wt.Path)]
		out = append(out, WorktreeInventoryRow{
			Path:       wt.Path,
			Branch:     wt.Branch,
			IsMain:     wt.IsMain,
			Detached:   strings.TrimSpace(wt.Branch) == "",
			Registered: isRegistered,
			Locked:     wt.Locked,
			Prunable:   wt.Prunable,
		})
	}
	return out, nil
}

// FindWorktreeInInventory returns the inventory row for an absolute worktree path.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by RepoWorktreeInventory."
func FindWorktreeInInventory(rows []WorktreeInventoryRow, path string) (*WorktreeInventoryRow, bool) {
	want := worktreePathKey(path)
	for i := range rows {
		if worktreePathKey(rows[i].Path) == want {
			return &rows[i], true
		}
	}
	return nil, false
}

// WorktreeCheckoutStatusRow is live checkout git state for one registered worktree.
type WorktreeCheckoutStatusRow = contract.WorktreeCheckoutStatusRow

const worktreeCheckoutStatusParallel = 4

// RepoWorktreeCheckoutStatus reads git checkout state for branch-bound worktrees in a repository.
func (s *Store) RepoWorktreeCheckoutStatus(
	ctx context.Context,
	repo gitdomain.GitRepository,
	gitSvc gitwork.Service,
) ([]WorktreeCheckoutStatusRow, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.RepoWorktreeCheckoutStatus")
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	registered, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		return nil, err
	}
	filtered := make([]gitdomain.GitWorktree, 0, len(registered))
	for _, wt := range registered {
		if gitWorktreeIsFullyRegistered(wt) {
			filtered = append(filtered, wt)
		}
	}
	if len(filtered) == 0 {
		return []WorktreeCheckoutStatusRow{}, nil
	}
	if _, err := gitSvc.OpenRepository(ctx, repo.Path); err != nil {
		return nil, fmt.Errorf("open repository: %w", err)
	}

	out := make([]WorktreeCheckoutStatusRow, len(filtered))
	var wg sync.WaitGroup
	sem := make(chan struct{}, worktreeCheckoutStatusParallel)
	for i, wt := range filtered {
		wg.Add(1)
		go func(i int, wt gitdomain.GitWorktree) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out[i] = s.checkoutStatusForWorktree(ctx, wt, gitSvc)
		}(i, wt)
	}
	wg.Wait()
	return out, nil
}

func (s *Store) checkoutStatusForWorktree(
	ctx context.Context,
	wt gitdomain.GitWorktree,
	gitSvc gitwork.Service,
) WorktreeCheckoutStatusRow {
	row := WorktreeCheckoutStatusRow{WorktreeID: wt.ID}
	if _, err := os.Stat(wt.Path); err != nil {
		if os.IsNotExist(err) {
			row.Reason = "path_missing"
			return row
		}
		row.Reason = "git_error"
		return row
	}
	st, err := gitSvc.CheckoutStatus(ctx, wt.Path)
	if err != nil {
		slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.checkoutStatusForWorktree.err",
			"worktree_id", wt.ID, "path", wt.Path, "err", err)
		row.Reason = "git_error"
		return row
	}
	row.Available = true
	row.Status = st
	return row
}

// ProbeGitWorktree checks whether path is a linked worktree of the repository.
func (s *Store) ProbeGitWorktree(
	ctx context.Context,
	repoID, path string,
	gitSvc gitwork.Service,
) (GitWorktreeProbeResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.ProbeGitWorktree")
	path = strings.TrimSpace(path)
	if path == "" {
		return GitWorktreeProbeResult{}, fmt.Errorf("%w: path required", taskcoredomain.ErrInvalidInput)
	}
	repo, err := s.GetGitRepositoryByID(ctx, repoID)
	if err != nil {
		return GitWorktreeProbeResult{}, err
	}
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	belongs, err := gitSvc.BelongsToRepository(ctx, path, repo.Path)
	if err != nil {
		return GitWorktreeProbeResult{}, fmt.Errorf("belongs to repository: %w", err)
	}
	if !belongs {
		return GitWorktreeProbeResult{Path: filepath.Clean(path), Linked: false}, nil
	}
	inventory, err := s.RepoWorktreeInventory(ctx, repo, gitSvc)
	if err != nil {
		return GitWorktreeProbeResult{}, err
	}
	row, found := FindWorktreeInInventory(inventory, path)
	if !found {
		return GitWorktreeProbeResult{Path: filepath.Clean(path), Linked: false}, nil
	}
	return GitWorktreeProbeResult{
		Path:       row.Path,
		Linked:     true,
		IsMain:     row.IsMain,
		Branch:     row.Branch,
		Registered: row.Registered,
	}, nil
}
