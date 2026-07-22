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

func openUnlinkedProjectsDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := autoMigrateStoreModels(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func createRepo(t *testing.T, db *gorm.DB, now time.Time) string {
	t.Helper()
	id := uuid.NewString()
	path := t.TempDir()
	if err := db.Create(&gitmodel.GitRepository{
		ID:           id,
		Path:         path,
		GitCommonDir: path + "/.git",
		CreatedAt:    now,
		UpdatedAt:    now,
	}).Error; err != nil {
		t.Fatal(err)
	}
	return id
}

func createProject(t *testing.T, db *gorm.DB, p projectsdomain.Project) projectmodel.Project {
	t.Helper()
	row := projectmodel.FromDomainProject(p)
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	return row
}

func loadProject(t *testing.T, db *gorm.DB, id string) projectmodel.Project {
	t.Helper()
	var got projectmodel.Project
	if err := db.First(&got, "id = ?", id).Error; err != nil {
		t.Fatalf("project %s: %v", id, err)
	}
	return got
}

func TestMigrateUnlinkedProjects_soleRepoAttachesNonDefault(t *testing.T) {
	db := openUnlinkedProjectsDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	repoID := createRepo(t, db, now)

	linked := createProject(t, db, projectsdomain.Project{
		ID: uuid.NewString(), Name: projectsdomain.DefaultProjectName,
		Status: projectsdomain.ProjectStatusActive, RepositoryID: &repoID,
		IsDefault: true, CreatedAt: now, UpdatedAt: now,
	})
	unlinked := createProject(t, db, projectsdomain.Project{
		ID: uuid.NewString(), Name: "Payment Improvements",
		Status: projectsdomain.ProjectStatusActive, IsDefault: false,
		CreatedAt: now, UpdatedAt: now,
	})

	if err := migrateUnlinkedProjects(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := migrateUnlinkedProjects(ctx, db); err != nil {
		t.Fatal(err)
	}

	got := loadProject(t, db, unlinked.ID)
	if got.RepositoryID == nil || *got.RepositoryID != repoID {
		t.Fatalf("unlinked repository_id=%v want %s", got.RepositoryID, repoID)
	}
	got = loadProject(t, db, linked.ID)
	if got.RepositoryID == nil || *got.RepositoryID != repoID {
		t.Fatalf("linked default repository_id=%v want %s", got.RepositoryID, repoID)
	}
}

func TestMigrateUnlinkedProjects_twoReposLeavesUnlinked(t *testing.T) {
	db := openUnlinkedProjectsDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	_ = createRepo(t, db, now)
	_ = createRepo(t, db, now)

	unlinked := createProject(t, db, projectsdomain.Project{
		ID: uuid.NewString(), Name: "Orphan",
		Status: projectsdomain.ProjectStatusActive, IsDefault: false,
		CreatedAt: now, UpdatedAt: now,
	})

	if err := migrateUnlinkedProjects(ctx, db); err != nil {
		t.Fatal(err)
	}

	got := loadProject(t, db, unlinked.ID)
	if got.RepositoryID != nil {
		t.Fatalf("repository_id=%v want nil", got.RepositoryID)
	}
}

func TestMigrateUnlinkedProjects_zeroReposLeavesUnlinked(t *testing.T) {
	db := openUnlinkedProjectsDB(t)
	ctx := context.Background()
	now := time.Now().UTC()

	unlinked := createProject(t, db, projectsdomain.Project{
		ID: uuid.NewString(), Name: "Orphan",
		Status: projectsdomain.ProjectStatusActive, IsDefault: false,
		CreatedAt: now, UpdatedAt: now,
	})

	if err := migrateUnlinkedProjects(ctx, db); err != nil {
		t.Fatal(err)
	}

	got := loadProject(t, db, unlinked.ID)
	if got.RepositoryID != nil {
		t.Fatalf("repository_id=%v want nil", got.RepositoryID)
	}
}

func TestMigrateUnlinkedProjects_nullDefaultUntouched(t *testing.T) {
	db := openUnlinkedProjectsDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	_ = createRepo(t, db, now)

	legacyDefault := createProject(t, db, projectsdomain.Project{
		ID: uuid.NewString(), Name: projectsdomain.DefaultProjectName,
		Status: projectsdomain.ProjectStatusActive, IsDefault: true,
		CreatedAt: now, UpdatedAt: now,
	})

	if err := migrateUnlinkedProjects(ctx, db); err != nil {
		t.Fatal(err)
	}

	got := loadProject(t, db, legacyDefault.ID)
	if got.RepositoryID != nil {
		t.Fatalf("default repository_id=%v want nil", got.RepositoryID)
	}
}
