package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"errors"
	"fmt"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ListAllGitRepositories returns every registered repository ordered by created_at.
func (s *Store) ListAllGitRepositories(ctx context.Context) ([]gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.ListAllGitRepositories")
	var rows []model.GitRepository
	err := s.db.WithContext(ctx).Order("created_at ASC").Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list all git repositories: %w", err)
	}
	return model.ToDomainGitRepositories(rows), nil
}

// GitRepositoryListSummary augments a repository row with list-page metadata.
type GitRepositoryListSummary = contract.GitRepositoryListSummary

// ListAllGitRepositoriesWithSummary returns repositories with main-branch name and
// linked worktree counts for the global list UI (mirrors web isLinkedWorktreeForDisplay).
func (s *Store) ListAllGitRepositoriesWithSummary(ctx context.Context) ([]GitRepositoryListSummary, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.ListAllGitRepositoriesWithSummary")
	repos, err := s.ListAllGitRepositories(ctx)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		return nil, nil
	}
	repoIDs := make([]string, len(repos))
	for i, repo := range repos {
		repoIDs[i] = repo.ID
	}

	type mainBranchRow struct {
		RepositoryID string
		Name         string
	}
	var mainBranches []mainBranchRow
	err = s.db.WithContext(ctx).
		Table("git_worktrees AS w").
		Select("w.repository_id, b.name").
		Joins("JOIN git_branches AS b ON b.id = w.branch_id").
		Where("w.is_main = ? AND w.repository_id IN ?", true, repoIDs).
		Scan(&mainBranches).Error
	if err != nil {
		return nil, fmt.Errorf("list main branch names: %w", err)
	}
	mainBranchByRepo := make(map[string]string, len(mainBranches))
	for _, row := range mainBranches {
		mainBranchByRepo[row.RepositoryID] = row.Name
	}

	type worktreeCountRow struct {
		RepositoryID string
		Count        int64
	}
	var worktreeCounts []worktreeCountRow
	err = s.db.WithContext(ctx).
		Model(&model.GitWorktree{}).
		Select("repository_id, COUNT(*) AS count").
		Where("is_main = ? AND branch_id <> '' AND repository_id IN ?", false, repoIDs).
		Group("repository_id").
		Scan(&worktreeCounts).Error
	if err != nil {
		return nil, fmt.Errorf("count linked worktrees: %w", err)
	}
	countByRepo := make(map[string]int, len(worktreeCounts))
	for _, row := range worktreeCounts {
		countByRepo[row.RepositoryID] = int(row.Count)
	}

	out := make([]GitRepositoryListSummary, len(repos))
	for i, repo := range repos {
		out[i] = GitRepositoryListSummary{
			Repository:          repo,
			MainBranchName:      mainBranchByRepo[repo.ID],
			LinkedWorktreeCount: countByRepo[repo.ID],
		}
	}
	return out, nil
}

// CreateGlobalGitRepository registers a main checkout without project scoping and
// seeds the main worktree row with the checkout's current branch.
func (s *Store) CreateGlobalGitRepository(ctx context.Context, input CreateGitRepositoryInput, gitSvc gitwork.Service) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.CreateGlobalGitRepository")
	repo, err := s.registerGitRepository(ctx, input, gitSvc)
	if err != nil {
		return gitdomain.GitRepository{}, err
	}
	if err := s.seedMainWorktreeWithCurrentBranch(ctx, repo, gitSvc); err != nil {
		return gitdomain.GitRepository{}, err
	}
	return repo, nil
}

// DeleteGlobalGitRepository removes a repository by id when no running tasks reference it.
func (s *Store) DeleteGlobalGitRepository(ctx context.Context, repoID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.DeleteGlobalGitRepository")
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		return fmt.Errorf("%w: repository_id required", taskcoredomain.ErrInvalidInput)
	}
	if _, err := s.GetGitRepositoryByID(ctx, repoID); err != nil {
		return err
	}
	if err := guardNoRunningTask(ctx, s.db, repoID); err != nil {
		return err
	}
	res := s.db.WithContext(ctx).Delete(&model.GitRepository{}, "id = ?", repoID)
	if res.Error != nil {
		return fmt.Errorf("delete git repository: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gitdomain.NewGitErr(gitdomain.GitCodeRepositoryNotFound, "repository not found")
	}
	return nil
}

// ListGitWorktreesByRepo returns worktrees for a repository (no project scope).
func (s *Store) ListGitWorktreesByRepo(ctx context.Context, repoID string) ([]gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.ListGitWorktreesByRepo")
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		return nil, fmt.Errorf("%w: repository_id required", taskcoredomain.ErrInvalidInput)
	}
	if _, err := s.GetGitRepositoryByID(ctx, repoID); err != nil {
		return nil, err
	}
	var rows []model.GitWorktree
	err := s.db.WithContext(ctx).
		Where("repository_id = ?", repoID).
		Order("is_main DESC, created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list git worktrees: %w", err)
	}
	return model.ToDomainGitWorktrees(rows), nil
}

// CreateGitWorktreeForRepo adds a linked worktree via git under a repository.
func (s *Store) CreateGitWorktreeForRepo(ctx context.Context, repoID string, input CreateGitWorktreeInput, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.CreateGitWorktreeForRepo")
	repo, err := s.GetGitRepositoryByID(ctx, repoID)
	if err != nil {
		return gitdomain.GitWorktree{}, err
	}
	return s.createGitWorktreeOnRepo(ctx, repo, input, gitSvc)
}

// RegisterExistingGitWorktree validates path is a linked worktree of repo, inserts a row,
// and optionally binds a branch association in the same flow.
func (s *Store) RegisterExistingGitWorktree(
	ctx context.Context,
	repoID string,
	path, name string,
	bind BindBranchInput,
	gitSvc gitwork.Service,
) (gitdomain.GitWorktree, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.RegisterExistingGitWorktree")
	repo, err := s.GetGitRepositoryByID(ctx, repoID)
	if err != nil {
		return gitdomain.GitWorktree{}, err
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return gitdomain.GitWorktree{}, fmt.Errorf("%w: path required", taskcoredomain.ErrInvalidInput)
	}
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	inventory, err := s.RepoWorktreeInventory(ctx, repo, gitSvc)
	if err != nil {
		return gitdomain.GitWorktree{}, err
	}
	invRow, ok := FindWorktreeInInventory(inventory, path)
	if !ok {
		return gitdomain.GitWorktree{}, fmt.Errorf("%w: path is not a linked worktree of this repository", taskcoredomain.ErrInvalidInput)
	}
	if invRow.Registered {
		return gitdomain.GitWorktree{}, gitdomain.NewGitErr(gitdomain.GitCodePathExists, "worktree path already registered")
	}
	label := strings.TrimSpace(name)
	if label == "" {
		label = worktreeDisplayName(invRow.Path)
	}
	bindName := strings.TrimSpace(bind.Name)
	if bindName == "" {
		bindName = strings.TrimSpace(invRow.Branch)
	}
	if bindName == "" {
		return gitdomain.GitWorktree{}, fmt.Errorf("%w: branch required", taskcoredomain.ErrInvalidInput)
	}
	br, err := s.resolveBranchForWorktree(ctx, repo, "", BindBranchInput{
		Name:         bindName,
		CreateBranch: bind.CreateBranch,
		StartPoint:   bind.StartPoint,
	}, gitSvc)
	if err != nil {
		return gitdomain.GitWorktree{}, err
	}
	if existing, found, err := s.findGitWorktreeByRepoPath(ctx, repo.ID, invRow.Path); err != nil {
		return gitdomain.GitWorktree{}, err
	} else if found && !gitWorktreeIsFullyRegistered(existing) {
		if err := s.db.WithContext(ctx).Model(&model.GitWorktree{}).Where("id = ?", existing.ID).Updates(map[string]any{
			"name":      label,
			"branch_id": br.ID,
			"is_main":   invRow.IsMain,
		}).Error; err != nil {
			return gitdomain.GitWorktree{}, fmt.Errorf("complete git worktree registration: %w", err)
		}
		existing.Name = label
		existing.BranchID = br.ID
		existing.IsMain = invRow.IsMain
		return existing, nil
	}
	now := time.Now().UTC()
	wt := gitdomain.GitWorktree{
		ID:           uuid.NewString(),
		RepositoryID: repo.ID,
		Path:         invRow.Path,
		Name:         label,
		IsMain:       invRow.IsMain,
		BranchID:     br.ID,
		CreatedAt:    now,
	}
	wtRow := model.FromDomainGitWorktree(wt)
	if err := s.db.WithContext(ctx).Create(&wtRow).Error; err != nil {
		if storekernel.IsDuplicateKey(err) {
			return gitdomain.GitWorktree{}, gitdomain.NewGitErr(gitdomain.GitCodePathExists, "worktree path already registered")
		}
		return gitdomain.GitWorktree{}, fmt.Errorf("register git worktree: %w", err)
	}
	return wt, nil
}

// UnregisterGitWorktreeByID removes Hamix registration for a worktree without
// running git worktree remove — the checkout directory stays on disk.
func (s *Store) UnregisterGitWorktreeByID(ctx context.Context, worktreeID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.UnregisterGitWorktreeByID")
	if _, err := s.GetGitWorktreeByID(ctx, worktreeID); err != nil {
		return err
	}
	if err := guardNoRunningTask(ctx, s.db, worktreeID); err != nil {
		return err
	}
	res := s.db.WithContext(ctx).Delete(&model.GitWorktree{}, "id = ?", worktreeID)
	if res.Error != nil {
		return fmt.Errorf("unregister git worktree row: %w", res.Error)
	}
	return nil
}

// ListGitBranchesByRepo returns branches for a repository (no project scope).
func (s *Store) ListGitBranchesByRepo(ctx context.Context, repoID string) ([]gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.ListGitBranchesByRepo")
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		return nil, fmt.Errorf("%w: repository_id required", taskcoredomain.ErrInvalidInput)
	}
	if _, err := s.GetGitRepositoryByID(ctx, repoID); err != nil {
		return nil, err
	}
	var rows []model.GitBranch
	err := s.db.WithContext(ctx).
		Where("repository_id = ?", repoID).
		Order("name ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list git branches: %w", err)
	}
	return model.ToDomainGitBranches(rows), nil
}

// CreateGitBranchForRepo creates a branch via git under a repository (no project scope).
func (s *Store) CreateGitBranchForRepo(ctx context.Context, repoID string, input CreateGitBranchInput, gitSvc gitwork.Service) (gitdomain.GitBranch, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.CreateGitBranchForRepo")
	repo, err := s.GetGitRepositoryByID(ctx, repoID)
	if err != nil {
		return gitdomain.GitBranch{}, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return gitdomain.GitBranch{}, fmt.Errorf("%w: name required", taskcoredomain.ErrInvalidInput)
	}
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	opened, err := gitSvc.OpenRepository(ctx, repo.Path)
	if err != nil {
		return gitdomain.GitBranch{}, fmt.Errorf("open repository: %w", err)
	}
	created, err := gitSvc.CreateBranch(ctx, opened, name, strings.TrimSpace(input.StartPoint))
	if err != nil {
		if errors.Is(err, gitwork.ErrBranchExists) {
			return gitdomain.GitBranch{}, gitdomain.NewGitErr(gitdomain.GitCodeBranchExists, "branch already exists")
		}
		return gitdomain.GitBranch{}, err
	}
	row := gitdomain.GitBranch{
		ID:           uuid.NewString(),
		RepositoryID: repo.ID,
		Name:         created.Name,
		HeadSHA:      created.HeadSHA,
		CreatedAt:    time.Now().UTC(),
	}
	branchRow := model.FromDomainGitBranch(row)
	if err := s.db.WithContext(ctx).Create(&branchRow).Error; err != nil {
		if storekernel.IsDuplicateKey(err) {
			return gitdomain.GitBranch{}, gitdomain.NewGitErr(gitdomain.GitCodeBranchExists, "branch already exists")
		}
		return gitdomain.GitBranch{}, fmt.Errorf("create git branch row: %w", err)
	}
	return row, nil
}

//funclogmeasure:skip category=hot-path reason="Internal helper; trace emitted by calling chokepoint."
func (s *Store) createGitWorktreeOnRepo(ctx context.Context, repo gitdomain.GitRepository, input CreateGitWorktreeInput, gitSvc gitwork.Service) (gitdomain.GitWorktree, error) {
	path := strings.TrimSpace(input.Path)
	branch := strings.TrimSpace(input.Branch)
	if path == "" || branch == "" {
		return gitdomain.GitWorktree{}, fmt.Errorf("%w: path and branch required", taskcoredomain.ErrInvalidInput)
	}
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	if br, lookupErr := s.ResolveOrCreateBranchForRepo(ctx, repo, BindBranchInput{
		Name: branch, CreateBranch: false,
	}, gitSvc); lookupErr == nil {
		if err := s.GuardBranchNotBoundToOtherWorktree(ctx, br.ID, ""); err != nil {
			return gitdomain.GitWorktree{}, err
		}
	}
	opened, err := gitSvc.OpenRepository(ctx, repo.Path)
	if err != nil {
		return gitdomain.GitWorktree{}, fmt.Errorf("open repository: %w", err)
	}
	wt, err := gitSvc.AddWorktree(ctx, opened, path, gitwork.AddWorktreeOptions{
		Branch:       branch,
		CreateBranch: input.CreateBranch,
		StartPoint:   strings.TrimSpace(input.StartPoint),
	})
	if err != nil {
		return gitdomain.GitWorktree{}, mapGitworkCreateErr(err)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = worktreeDisplayName(wt.Path)
	}
	br, err := s.resolveBranchForWorktree(ctx, repo, "", BindBranchInput{
		Name:         branch,
		CreateBranch: false,
		StartPoint:   strings.TrimSpace(input.StartPoint),
	}, gitSvc)
	if err != nil {
		_ = gitSvc.RemoveWorktree(ctx, opened, wt.Path, true)
		return gitdomain.GitWorktree{}, err
	}
	now := time.Now().UTC()
	row := gitdomain.GitWorktree{
		ID:           uuid.NewString(),
		RepositoryID: repo.ID,
		Path:         wt.Path,
		Name:         name,
		IsMain:       false,
		BranchID:     br.ID,
		CreatedAt:    now,
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wtRow := model.FromDomainGitWorktree(row)
		if err := tx.Create(&wtRow).Error; err != nil {
			if storekernel.IsDuplicateKey(err) {
				return gitdomain.NewGitErr(gitdomain.GitCodePathExists, "worktree path already registered")
			}
			return err
		}
		return nil
	})
	if err != nil {
		_ = gitSvc.RemoveWorktree(ctx, opened, wt.Path, true)
		return gitdomain.GitWorktree{}, err
	}
	return row, nil
}
