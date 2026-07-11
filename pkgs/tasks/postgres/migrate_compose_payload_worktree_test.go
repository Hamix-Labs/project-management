package postgres

import (
	"context"
	"encoding/json"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"testing"
	"time"

	gitmodel "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMigrateComposePayloadWorktree_backfillsTemplatePayload(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC()

	if err := model.AutoMigrateAll(db); err != nil {
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
	brID := "br-1"
	extraBrID := "br-2"
	mainWtID := "wt-main"
	extraWtID := "wt-extra"
	br := gitmodel.FromDomainGitBranch(gitdomain.GitBranch{
		ID: brID, RepositoryID: repo.ID, Name: "main", CreatedAt: now,
	})
	if err := db.WithContext(ctx).Create(&br).Error; err != nil {
		t.Fatal(err)
	}
	extraBr := gitmodel.FromDomainGitBranch(gitdomain.GitBranch{
		ID: extraBrID, RepositoryID: repo.ID, Name: "feature", CreatedAt: now,
	})
	if err := db.WithContext(ctx).Create(&extraBr).Error; err != nil {
		t.Fatal(err)
	}
	mainWt := gitmodel.FromDomainGitWorktree(gitdomain.GitWorktree{
		ID: mainWtID, RepositoryID: repo.ID, Path: repo.Path, Name: "main", IsMain: true, BranchID: brID, CreatedAt: now,
	})
	if err := db.WithContext(ctx).Create(&mainWt).Error; err != nil {
		t.Fatal(err)
	}
	extraWt := gitmodel.FromDomainGitWorktree(gitdomain.GitWorktree{
		ID: extraWtID, RepositoryID: repo.ID, Path: "/repos/app-extra", Name: "extra", BranchID: extraBrID, CreatedAt: now.Add(time.Minute),
	})
	if err := db.WithContext(ctx).Create(&extraWt).Error; err != nil {
		t.Fatal(err)
	}
	repoID := repo.ID
	proj := projectmodel.FromDomainProject(projectsdomain.Project{
		ID:           "proj-1",
		Name:         "Default",
		Status:       projectsdomain.ProjectStatusActive,
		RepositoryID: &repoID,
		IsDefault:    true,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err := db.WithContext(ctx).Create(&proj).Error; err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(map[string]any{
		"title":      "Legacy template",
		"priority":   "medium",
		"status":     "ready",
		"project_id": proj.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	tmpl := model.TaskTemplate{
		ID:          "tmpl-1",
		Name:        "Legacy template",
		PayloadJSON: payload,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.WithContext(ctx).Create(&tmpl).Error; err != nil {
		t.Fatal(err)
	}

	if err := migrateComposePayloadWorktree(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := migrateComposePayloadWorktree(ctx, db); err != nil {
		t.Fatal("second migrate run should be idempotent")
	}

	var got model.TaskTemplate
	if err := db.WithContext(ctx).First(&got, "id = ?", tmpl.ID).Error; err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(got.PayloadJSON, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["worktree_id"] != mainWtID {
		t.Fatalf("worktree_id = %v, want %s (main worktree)", decoded["worktree_id"], mainWtID)
	}
}

func TestMigrateComposePayloadWorktree_remapsLegacyGlobalDefaultProject(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC()

	if err := model.AutoMigrateAll(db); err != nil {
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
	brID := "br-1"
	mainWtID := "wt-main"
	br := gitmodel.FromDomainGitBranch(gitdomain.GitBranch{
		ID: brID, RepositoryID: repo.ID, Name: "main", CreatedAt: now,
	})
	if err := db.WithContext(ctx).Create(&br).Error; err != nil {
		t.Fatal(err)
	}
	mainWt := gitmodel.FromDomainGitWorktree(gitdomain.GitWorktree{
		ID: mainWtID, RepositoryID: repo.ID, Path: repo.Path, Name: "main", IsMain: true, BranchID: brID, CreatedAt: now,
	})
	if err := db.WithContext(ctx).Create(&mainWt).Error; err != nil {
		t.Fatal(err)
	}
	repoID := repo.ID
	proj := projectmodel.FromDomainProject(projectsdomain.Project{
		ID:           "proj-default",
		Name:         "Default",
		Status:       projectsdomain.ProjectStatusActive,
		RepositoryID: &repoID,
		IsDefault:    true,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err := db.WithContext(ctx).Create(&proj).Error; err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(map[string]any{
		"title":      "Legacy template",
		"priority":   "medium",
		"status":     "ready",
		"project_id": projectsdomain.LegacyGlobalDefaultProjectID,
	})
	if err != nil {
		t.Fatal(err)
	}
	tmpl := model.TaskTemplate{
		ID:          "tmpl-legacy",
		Name:        "Legacy template",
		PayloadJSON: payload,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.WithContext(ctx).Create(&tmpl).Error; err != nil {
		t.Fatal(err)
	}

	if err := migrateComposePayloadWorktree(ctx, db); err != nil {
		t.Fatal(err)
	}

	var got model.TaskTemplate
	if err := db.WithContext(ctx).First(&got, "id = ?", tmpl.ID).Error; err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(got.PayloadJSON, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["worktree_id"] != mainWtID {
		t.Fatalf("worktree_id = %v, want %s", decoded["worktree_id"], mainWtID)
	}
	if decoded["project_id"] != proj.ID {
		t.Fatalf("project_id = %v, want %s", decoded["project_id"], proj.ID)
	}
}
