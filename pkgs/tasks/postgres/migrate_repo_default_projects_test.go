package postgres

import (
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"context"
	"testing"
	"time"

	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	gitmodel "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMigrateRepoDefaultProjects_removesGlobalDefault(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC()

	if err := model.AutoMigrateAll(db); err != nil {
		t.Fatal(err)
	}

	legacy := projectmodel.FromDomainProject(projectsdomain.LegacyGlobalDefaultProject(now))
	if err := db.WithContext(ctx).Create(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	repo := gitmodel.FromDomainGitRepository(gitdomain.GitRepository{
		ID:            "repo-1",
		Path:          "/repos/app",
		GitCommonDir:  "/repos/app/.git",
		DefaultBranch: "main",
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err := db.WithContext(ctx).Create(&repo).Error; err != nil {
		t.Fatal(err)
	}
	wtID := "wt-1"
	brID := "br-1"
	br := gitmodel.FromDomainGitBranch(gitdomain.GitBranch{
		ID: brID, RepositoryID: repo.ID, Name: "main", CreatedAt: now,
	})
	if err := db.WithContext(ctx).Create(&br).Error; err != nil {
		t.Fatal(err)
	}
	wt := gitmodel.FromDomainGitWorktree(gitdomain.GitWorktree{
		ID: wtID, RepositoryID: repo.ID, Path: repo.Path, Name: "main", IsMain: true, BranchID: brID, CreatedAt: now,
	})
	if err := db.WithContext(ctx).Create(&wt).Error; err != nil {
		t.Fatal(err)
	}
	legacyID := projectsdomain.LegacyGlobalDefaultProjectID
	task := model.FromDomainTask(domain.Task{
		ID: "task-1", Title: "t", InitialPrompt: "p", Status: domain.StatusReady, Priority: domain.PriorityMedium,
		Runner: "cursor", ProjectID: &legacyID, WorktreeID: &wtID,
	})
	if err := db.WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatal(err)
	}

	if err := migrateRepoDefaultProjects(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := migrateRepoDefaultProjects(ctx, db); err != nil {
		t.Fatal("second migrate run should be idempotent")
	}

	var globalCount int64
	if err := db.WithContext(ctx).Model(&projectmodel.Project{}).
		Where("id = ?", projectsdomain.LegacyGlobalDefaultProjectID).
		Count(&globalCount).Error; err != nil {
		t.Fatal(err)
	}
	if globalCount != 0 {
		t.Fatalf("global default row count = %d, want 0", globalCount)
	}

	var defaultProjRow projectmodel.Project
	if err := db.WithContext(ctx).
		Where("repository_id = ? AND is_default = ?", repo.ID, true).
		First(&defaultProjRow).Error; err != nil {
		t.Fatal(err)
	}
	defaultProj := projectmodel.ToDomainProject(defaultProjRow)
	taskRow := model.Task{}
	if err := db.WithContext(ctx).First(&taskRow, "id = ?", "task-1").Error; err != nil {
		t.Fatal(err)
	}
	gotTask := model.ToDomainTask(taskRow)
	if gotTask.ProjectID == nil || *gotTask.ProjectID != defaultProj.ID {
		t.Fatalf("task project_id = %v, want %s", gotTask.ProjectID, defaultProj.ID)
	}
}
