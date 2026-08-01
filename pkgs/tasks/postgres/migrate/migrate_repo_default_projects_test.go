package migrate

import (
	"context"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"testing"
	"time"

	gitmodel "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcoremodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
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

	if err := autoMigrateStoreModels(db); err != nil {
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
	task := taskcoremodel.FromDomainTask(taskcoredomain.Task{
		ID: "task-1", Title: "t", InitialPrompt: "p", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
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
	taskRow := taskcoremodel.Task{}
	if err := db.WithContext(ctx).First(&taskRow, "id = ?", "task-1").Error; err != nil {
		t.Fatal(err)
	}
	gotTask := taskcoremodel.ToDomainTask(taskRow)
	if gotTask.ProjectID == nil || *gotTask.ProjectID != defaultProj.ID {
		t.Fatalf("task project_id = %v, want %s", gotTask.ProjectID, defaultProj.ID)
	}
}

func TestMigrateRepoDefaultProjects_skipsWhenGlobalDefaultExists(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC()
	if err := autoMigrateStoreModels(db); err != nil {
		t.Fatal(err)
	}

	repo := gitmodel.FromDomainGitRepository(gitdomain.GitRepository{
		ID: "repo-1", Path: "/repos/app", GitCommonDir: "/repos/app/.git",
		DefaultBranch: "main", CreatedAt: now, UpdatedAt: now,
	})
	if err := db.WithContext(ctx).Create(&repo).Error; err != nil {
		t.Fatal(err)
	}
	global := projectmodel.FromDomainProject(projectsdomain.GlobalDefaultProject(now))
	if err := db.WithContext(ctx).Create(&global).Error; err != nil {
		t.Fatal(err)
	}

	if err := migrateRepoDefaultProjects(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := migrateRepoDefaultProjects(ctx, db); err != nil {
		t.Fatal("second migrate run should be idempotent:", err)
	}

	var defaults []projectmodel.Project
	if err := db.WithContext(ctx).Where("is_default = ?", true).Find(&defaults).Error; err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 1 {
		t.Fatalf("default count=%d want 1: %#v", len(defaults), defaults)
	}
	if defaults[0].ID != projectsdomain.GlobalDefaultProjectID {
		t.Fatalf("default id=%s want %s", defaults[0].ID, projectsdomain.GlobalDefaultProjectID)
	}
	if defaults[0].RepositoryID != nil {
		t.Fatalf("global default got repository_id=%v", defaults[0].RepositoryID)
	}

	var repoDefaults int64
	if err := db.WithContext(ctx).Model(&projectmodel.Project{}).
		Where("repository_id = ? AND is_default = ?", repo.ID, true).
		Count(&repoDefaults).Error; err != nil {
		t.Fatal(err)
	}
	if repoDefaults != 0 {
		t.Fatalf("per-repo default count=%d want 0 after ADR-0094 skip", repoDefaults)
	}
}

func TestMigrateRun_idempotentAfterGlobalDefault(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC()
	if err := autoMigrateStoreModels(db); err != nil {
		t.Fatal(err)
	}

	repo := gitmodel.FromDomainGitRepository(gitdomain.GitRepository{
		ID: "repo-a", Path: "/a", GitCommonDir: "/a/.git",
		DefaultBranch: "main", CreatedAt: now, UpdatedAt: now,
	})
	if err := db.Create(&repo).Error; err != nil {
		t.Fatal(err)
	}
	repoID := repo.ID
	def := projectmodel.Project{
		ID: "def-a", Name: "Default", Status: projectsdomain.ProjectStatusActive,
		RepositoryID: &repoID, IsDefault: true, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&def).Error; err != nil {
		t.Fatal(err)
	}

	if err := migrateRepoDefaultProjects(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := migrateGlobalDefaultProject(ctx, db); err != nil {
		t.Fatal(err)
	}
	// Simulates every later Migrate: rev-6 step runs again before rev-28.
	if err := migrateRepoDefaultProjects(ctx, db); err != nil {
		t.Fatal("repo default after global consolidate:", err)
	}
	if err := migrateGlobalDefaultProject(ctx, db); err != nil {
		t.Fatal("global default second run:", err)
	}

	var defaults []projectmodel.Project
	if err := db.Where("is_default = ?", true).Find(&defaults).Error; err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 1 {
		t.Fatalf("default count=%d want 1", len(defaults))
	}
	if defaults[0].RepositoryID != nil {
		t.Fatalf("still repo-bound: %v", defaults[0].RepositoryID)
	}
}
