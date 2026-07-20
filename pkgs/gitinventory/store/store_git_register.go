package store

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"fmt"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	projectsstore "github.com/AlexsanderHamir/Hamix/pkgs/projects/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// registerGitRepository resolves git identity and inserts a repository row.
// Duplicate detection uses git_common_dir, not path.
func (s *Store) registerGitRepository(ctx context.Context, input CreateGitRepositoryInput, gitSvc gitwork.Service) (gitdomain.GitRepository, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.registerGitRepository")
	path := strings.TrimSpace(input.Path)
	if path == "" {
		return gitdomain.GitRepository{}, fmt.Errorf("%w: path required", taskcoredomain.ErrInvalidInput)
	}
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	mainRoot, commonDir, err := gitSvc.ResolveRegistration(ctx, path)
	if err != nil {
		if errors.Is(err, gitwork.ErrNotARepository) {
			return gitdomain.GitRepository{}, gitdomain.NewGitErr(gitdomain.GitCodeNotARepository, "path is not a git repository")
		}
		return gitdomain.GitRepository{}, fmt.Errorf("resolve repository: %w", err)
	}
	var existing int64
	if err := s.db.WithContext(ctx).Model(&model.GitRepository{}).
		Where("git_common_dir = ?", commonDir).
		Count(&existing).Error; err != nil {
		return gitdomain.GitRepository{}, err
	}
	if existing > 0 {
		return gitdomain.GitRepository{}, gitdomain.NewGitErr(gitdomain.GitCodeDuplicate, "repository already registered")
	}
	defaultBranch := strings.TrimSpace(input.DefaultBranch)
	if defaultBranch == "" {
		opened, openErr := gitSvc.OpenRepository(ctx, mainRoot)
		if openErr != nil {
			return gitdomain.GitRepository{}, fmt.Errorf("open repository for default branch: %w", openErr)
		}
		resolved, resolveErr := gitSvc.ResolveDefaultBranch(ctx, opened, "origin")
		if resolveErr != nil {
			return gitdomain.GitRepository{}, fmt.Errorf("resolve default branch: %w", resolveErr)
		}
		defaultBranch = strings.TrimSpace(resolved)
		if defaultBranch == "" {
			defaultBranch = "main"
		}
	}
	now := time.Now().UTC()
	repo := gitdomain.GitRepository{
		ID:            uuid.NewString(),
		Path:          mainRoot,
		GitCommonDir:  commonDir,
		HostPath:      strings.TrimSpace(input.HostPath),
		DefaultBranch: defaultBranch,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repoRow := model.FromDomainGitRepository(repo)
		if err := tx.Create(&repoRow).Error; err != nil {
			if storekernel.IsDuplicateKey(err) {
				return gitdomain.NewGitErr(gitdomain.GitCodeDuplicate, "repository already registered")
			}
			return err
		}
		if _, err := projectsstore.CreateDefaultProjectForRepo(ctx, tx, repo.ID, now); err != nil {
			return fmt.Errorf("seed default project: %w", err)
		}
		return nil
	})
	if err != nil {
		return gitdomain.GitRepository{}, err
	}
	return repo, nil
}

// seedMainWorktreeWithCurrentBranch inserts the main worktree row and one branch
// row for the checkout branch currently at the main root.
func (s *Store) seedMainWorktreeWithCurrentBranch(ctx context.Context, repo gitdomain.GitRepository, gitSvc gitwork.Service) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.seedMainWorktreeWithCurrentBranch")
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	opened, err := gitSvc.OpenRepository(ctx, repo.Path)
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	branches, err := gitSvc.ListBranches(ctx, opened)
	if err != nil {
		return fmt.Errorf("list branches: %w", err)
	}
	branchName := strings.TrimSpace(repo.DefaultBranch)
	if branchName == "" {
		branchName = "main"
	}
	headSHA := ""
	for _, b := range branches {
		if b.IsCurrent && strings.TrimSpace(b.Name) != "" {
			branchName = b.Name
			headSHA = b.HeadSHA
			break
		}
	}
	now := time.Now().UTC()
	branchRow := gitdomain.GitBranch{
		ID:           uuid.NewString(),
		RepositoryID: repo.ID,
		Name:         branchName,
		HeadSHA:      headSHA,
		CreatedAt:    now,
	}
	mainWT := gitdomain.GitWorktree{
		ID:           uuid.NewString(),
		RepositoryID: repo.ID,
		Path:         repo.Path,
		Name:         worktreeDisplayName(repo.Path),
		IsMain:       true,
		BranchID:     branchRow.ID,
		CreatedAt:    now,
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		branchModel := model.FromDomainGitBranch(branchRow)
		if err := tx.Create(&branchModel).Error; err != nil {
			return err
		}
		mainWTModel := model.FromDomainGitWorktree(mainWT)
		if err := tx.Create(&mainWTModel).Error; err != nil {
			return err
		}
		return nil
	})
}
