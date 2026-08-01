package migrate

import (
	"context"
	"testing"
	"time"

	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	gitmodel "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcoremodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMigrateGlobalDefaultProject_consolidatesPerRepoDefaults(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC()
	if err := autoMigrateStoreModels(db); err != nil {
		t.Fatal(err)
	}

	repoA := gitmodel.FromDomainGitRepository(gitdomain.GitRepository{
		ID: "repo-a", Path: "/a", GitCommonDir: "/a/.git", DefaultBranch: "main", CreatedAt: now, UpdatedAt: now,
	})
	repoB := gitmodel.FromDomainGitRepository(gitdomain.GitRepository{
		ID: "repo-b", Path: "/b", GitCommonDir: "/b/.git", DefaultBranch: "main", CreatedAt: now, UpdatedAt: now,
	})
	if err := db.Create(&repoA).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&repoB).Error; err != nil {
		t.Fatal(err)
	}
	repoAID, repoBID := repoA.ID, repoB.ID
	defA := projectmodel.Project{
		ID: "def-a", Name: "Default", Status: projectsdomain.ProjectStatusActive,
		RepositoryID: &repoAID, IsDefault: true, CreatedAt: now, UpdatedAt: now,
	}
	defB := projectmodel.Project{
		ID: "def-b", Name: "Default", Status: projectsdomain.ProjectStatusActive,
		RepositoryID: &repoBID, IsDefault: true, CreatedAt: now.Add(time.Second), UpdatedAt: now,
	}
	if err := db.Create(&defA).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&defB).Error; err != nil {
		t.Fatal(err)
	}
	pidA := defA.ID
	task := taskcoremodel.FromDomainTask(taskcoredomain.Task{
		ID: "task-1", Title: "t", InitialPrompt: "p",
		Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
		Runner: "cursor", ProjectID: &pidA,
	})
	if err := db.Create(&task).Error; err != nil {
		t.Fatal(err)
	}

	if err := migrateGlobalDefaultProject(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := migrateGlobalDefaultProject(ctx, db); err != nil {
		t.Fatal("second run should be idempotent:", err)
	}

	var defaults []projectmodel.Project
	if err := db.Where("is_default = ?", true).Find(&defaults).Error; err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 1 {
		t.Fatalf("default count=%d want 1: %#v", len(defaults), defaults)
	}
	if defaults[0].RepositoryID != nil {
		t.Fatalf("global default still has repository_id=%v", defaults[0].RepositoryID)
	}
	if defaults[0].Name != projectsdomain.DefaultProjectName {
		t.Fatalf("name=%q", defaults[0].Name)
	}

	var gotTask taskcoremodel.Task
	if err := db.First(&gotTask, "id = ?", "task-1").Error; err != nil {
		t.Fatal(err)
	}
	if gotTask.ProjectID == nil || *gotTask.ProjectID != defaults[0].ID {
		t.Fatalf("task project_id=%v want %s", gotTask.ProjectID, defaults[0].ID)
	}
	if gotTask.RepositoryID == nil || *gotTask.RepositoryID != repoAID {
		t.Fatalf("task repository_id=%v want %s", gotTask.RepositoryID, repoAID)
	}
}
