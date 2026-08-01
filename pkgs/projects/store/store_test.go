package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func newProjectStoreOnly(t *testing.T) (*Store, *gorm.DB, context.Context) {
	t.Helper()
	db := tasktestdb.OpenSQLite(t)
	return NewStore(db), db, context.Background()
}

func seedRepo(t *testing.T, db *gorm.DB, ctx context.Context) string {
	t.Helper()
	repoID := uuid.NewString()
	now := time.Now().UTC()
	path := t.TempDir()
	if err := db.WithContext(ctx).Table("git_repositories").Create(map[string]any{
		"id":             repoID,
		"path":           path,
		"git_common_dir": path,
		"host_path":      "",
		"default_branch": "main",
		"created_at":     now,
		"updated_at":     now,
	}).Error; err != nil {
		t.Fatalf("seed repo: %v", err)
	}
	return repoID
}

func TestProjectsStore_CreateProject_requiresRepositoryID(t *testing.T) {
	s, _, ctx := newProjectStoreOnly(t)
	if _, err := s.CreateProject(ctx, CreateProjectInput{Name: "No repo"}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("create without repo err = %v, want ErrInvalidInput", err)
	}
}

func TestProjectsStore_globalDefault_roundtrip(t *testing.T) {
	s, db, ctx := newProjectStoreOnly(t)
	defaultProj, err := EnsureGlobalDefaultProject(ctx, db, time.Now().UTC())
	if err != nil {
		t.Fatalf("EnsureGlobalDefaultProject: %v", err)
	}
	if !defaultProj.IsDefault || defaultProj.RepositoryID != nil {
		t.Fatalf("default project = %#v", defaultProj)
	}
	got, err := s.GetGlobalDefaultProject(ctx)
	if err != nil {
		t.Fatalf("GetGlobalDefaultProject: %v", err)
	}
	if got.ID != defaultProj.ID {
		t.Fatalf("got = %#v, want id %q", got, defaultProj.ID)
	}
	again, err := EnsureGlobalDefaultProject(ctx, db, time.Now().UTC())
	if err != nil {
		t.Fatalf("EnsureGlobalDefaultProject again: %v", err)
	}
	if again.ID != defaultProj.ID {
		t.Fatalf("idempotent ensure = %#v, want %#v", again, defaultProj)
	}
}

func TestProjectsStore_ListProjectsByRepository_includesGlobalDefault(t *testing.T) {
	s, db, ctx := newProjectStoreOnly(t)
	repoID := seedRepo(t, db, ctx)
	def, err := EnsureGlobalDefaultProject(ctx, db, time.Now().UTC())
	if err != nil {
		t.Fatalf("EnsureGlobalDefaultProject: %v", err)
	}
	custom, err := s.CreateProject(ctx, CreateProjectInput{
		Name:         "Custom",
		RepositoryID: &repoID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	rows, err := s.ListProjectsByRepository(ctx, repoID)
	if err != nil {
		t.Fatalf("ListProjectsByRepository: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("list len=%d want 2: %#v", len(rows), rows)
	}
	if !rows[0].IsDefault || rows[0].ID != def.ID {
		t.Fatalf("first row should be global default: %#v", rows[0])
	}
	if rows[1].ID != custom.ID {
		t.Fatalf("second row should be custom: %#v", rows[1])
	}
}

func TestProjectsStore_DeleteProjectsForRepository_keepsGlobalDefault(t *testing.T) {
	s, db, ctx := newProjectStoreOnly(t)
	repoID := seedRepo(t, db, ctx)
	defaultProj, err := EnsureGlobalDefaultProject(ctx, db, time.Now().UTC())
	if err != nil {
		t.Fatalf("EnsureGlobalDefaultProject: %v", err)
	}
	_, err = s.CreateProject(ctx, CreateProjectInput{
		Name:         "Custom",
		RepositoryID: &repoID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := DeleteProjectsForRepository(ctx, db, repoID); err != nil {
		t.Fatalf("DeleteProjectsForRepository: %v", err)
	}
	var n int64
	if err := db.WithContext(ctx).Model(&projectmodel.Project{}).Where("repository_id = ?", repoID).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("remaining repo projects=%d", n)
	}
	got, err := s.GetProject(ctx, defaultProj.ID)
	if err != nil {
		t.Fatalf("GetProject global default: %v", err)
	}
	if !got.IsDefault {
		t.Fatalf("global default missing: %#v", got)
	}
}
