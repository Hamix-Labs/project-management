package store

import (
	"context"
	"errors"
	"strings"
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

func seedRepoWithDefaultProject(t *testing.T, db *gorm.DB, ctx context.Context) (repoID string, defaultProj domain.Project) {
	t.Helper()
	repoID = uuid.NewString()
	now := time.Now().UTC()
	path := t.TempDir()
	// Insert by table/columns only — projects must not import gitinventory models.
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

func TestProjectsStore_CreateContext_rejectsOversizeTitleAndBody(t *testing.T) {
	s, db, ctx := newProjectStoreOnly(t)
	_, project := seedRepoWithDefaultProject(t, db, ctx)

	overTitle := strings.Repeat("t", domain.MaxProjectContextTitleChars+1)
	_, err := s.CreateProjectContext(ctx, project.ID, CreateProjectContextInput{
		Title: overTitle,
		Body:  "ok",
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("oversize title err = %v, want ErrInvalidInput", err)
	}

	overBody := strings.Repeat("b", domain.MaxProjectContextBodyBytes+1)
	_, err = s.CreateProjectContext(ctx, project.ID, CreateProjectContextInput{
		Title: "ok",
		Body:  overBody,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("oversize body err = %v, want ErrInvalidInput", err)
	}

	overDesc := strings.Repeat("d", domain.MaxProjectContextDescriptionChars+1)
	_, err = s.CreateProjectContext(ctx, project.ID, CreateProjectContextInput{
		Title:       "ok",
		Description: overDesc,
		Body:        "ok",
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("oversize description err = %v, want ErrInvalidInput", err)
	}

	item, err := s.CreateProjectContext(ctx, project.ID, CreateProjectContextInput{
		Tag:         "General",
		Title:       "alias",
		Description: "when to use this memory",
		Body:        "imported content",
	})
	if err != nil {
		t.Fatalf("valid create: %v", err)
	}
	if item.Description != "when to use this memory" {
		t.Fatalf("description = %q", item.Description)
	}

	_, err = s.UpdateProjectContext(ctx, project.ID, item.ID, UpdateProjectContextInput{
		Body: &overBody,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("oversize patch body err = %v, want ErrInvalidInput", err)
	}

	_, err = s.UpdateProjectContext(ctx, project.ID, item.ID, UpdateProjectContextInput{
		Description: &overDesc,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("oversize patch description err = %v, want ErrInvalidInput", err)
	}

	patchedDesc := "updated blurb"
	got, err := s.UpdateProjectContext(ctx, project.ID, item.ID, UpdateProjectContextInput{
		Description: &patchedDesc,
	})
	if err != nil {
		t.Fatalf("patch description: %v", err)
	}
	if got.Description != patchedDesc {
		t.Fatalf("patched description = %q", got.Description)
	}
}

func TestProjectsStore_DeleteProjectsForRepository_removesDefault(t *testing.T) {
	s, db, ctx := newProjectStoreOnly(t)
	repoID, defaultProj := seedRepoWithDefaultProject(t, db, ctx)
	_, err := s.CreateProject(ctx, CreateProjectInput{
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
		t.Fatalf("remaining projects=%d", n)
	}
	if _, err := s.GetProject(ctx, defaultProj.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetProject after delete: %v", err)
	}
}
