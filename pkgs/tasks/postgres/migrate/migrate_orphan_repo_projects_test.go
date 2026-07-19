package migrate

import (
	"context"
	"testing"
	"time"

	gitmodel "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMigrateOrphanRepoProjects_deletesDefaultsForMissingRepos(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC()
	if err := autoMigrateStoreModels(db); err != nil {
		t.Fatal(err)
	}

	liveRepoID := uuid.NewString()
	if err := db.WithContext(ctx).Create(&gitmodel.GitRepository{
		ID: liveRepoID, Path: t.TempDir(), CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatal(err)
	}
	liveDefault := projectmodel.FromDomainProject(projectsdomain.Project{
		ID: uuid.NewString(), Name: projectsdomain.DefaultProjectName,
		Description: "keep", Status: projectsdomain.ProjectStatusActive,
		RepositoryID: &liveRepoID, IsDefault: true, CreatedAt: now, UpdatedAt: now,
	})
	if err := db.WithContext(ctx).Create(&liveDefault).Error; err != nil {
		t.Fatal(err)
	}

	orphanRepoID := uuid.NewString()
	orphanDefault := projectmodel.FromDomainProject(projectsdomain.Project{
		ID: uuid.NewString(), Name: projectsdomain.DefaultProjectName,
		Description:  "Built-in project for tasks tied to this repository.",
		Status:       projectsdomain.ProjectStatusActive,
		RepositoryID: &orphanRepoID, IsDefault: true, CreatedAt: now, UpdatedAt: now,
	})
	if err := db.WithContext(ctx).Create(&orphanDefault).Error; err != nil {
		t.Fatal(err)
	}

	if err := migrateOrphanRepoProjects(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := migrateOrphanRepoProjects(ctx, db); err != nil {
		t.Fatal(err)
	}

	var n int64
	if err := db.WithContext(ctx).Model(&projectmodel.Project{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("projects count=%d want 1", n)
	}
	var kept projectmodel.Project
	if err := db.WithContext(ctx).First(&kept, "id = ?", liveDefault.ID).Error; err != nil {
		t.Fatalf("live default missing: %v", err)
	}
}
