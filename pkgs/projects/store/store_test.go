package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	gitmodel "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func newProjectStoreOnly(t *testing.T) (*Store, *gorm.DB, context.Context) {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	return NewStore(db), db, context.Background()
}

func seedRepoWithDefaultProject(t *testing.T, db *gorm.DB, ctx context.Context) (repoID string, defaultProj domain.Project) {
	t.Helper()
	repoID = uuid.NewString()
	now := time.Now().UTC()
	repo := gitmodel.GitRepository{
		ID:        repoID,
		Path:      t.TempDir(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.WithContext(ctx).Create(&repo).Error; err != nil {
		t.Fatalf("seed repo: %v", err)
	}
	defaultProj, err := CreateDefaultProjectForRepo(ctx, db, repoID, now)
	if err != nil {
		t.Fatalf("CreateDefaultProjectForRepo: %v", err)
	}
	return repoID, defaultProj
}

func TestProjectsStore_CreateProject_requiresRepositoryID(t *testing.T) {
	s, _, ctx := newProjectStoreOnly(t)
	if _, err := s.CreateProject(ctx, CreateProjectInput{Name: "No repo"}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("create without repo err = %v, want ErrInvalidInput", err)
	}
}

func TestProjectsStore_defaultProject_roundtrip(t *testing.T) {
	s, db, ctx := newProjectStoreOnly(t)
	repoID, defaultProj := seedRepoWithDefaultProject(t, db, ctx)
	if !defaultProj.IsDefault || defaultProj.RepositoryID == nil || *defaultProj.RepositoryID != repoID {
		t.Fatalf("default project = %#v", defaultProj)
	}
	got, err := s.GetDefaultProjectForRepository(ctx, repoID)
	if err != nil {
		t.Fatalf("GetDefaultProjectForRepository: %v", err)
	}
	if got.ID != defaultProj.ID {
		t.Fatalf("got = %#v, want id %q", got, defaultProj.ID)
	}
}
