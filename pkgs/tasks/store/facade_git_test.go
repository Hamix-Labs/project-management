package store

import (
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/model"
)

func gitTestStore(t *testing.T) (*Store, context.Context, gitwork.Service) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	return NewStore(tasktestdb.OpenSQLite(t)), context.Background(), gitwork.New()
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGitStore(t, dir, "init", "-b", "main")
	runGitStore(t, dir, "config", "user.email", "t@example.com")
	runGitStore(t, dir, "config", "user.name", "Test")
	runGitStore(t, dir, "commit", "--allow-empty", "-m", "init")
	return dir
}

func runGitStore(t *testing.T, dir string, args ...string) {
	t.Helper()
	all := append([]string{"-C", dir}, args...)
	if out, err := exec.Command("git", all...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestStore_GitRepositoryCRUD(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)

	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitRepository: %v", err)
	}
	list, err := s.ListGitRepositories(ctx, projectsdomain.LegacyGlobalDefaultProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("len=%d want 1", len(list))
	}
	got, err := s.GetGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Path == "" {
		t.Fatal("empty path")
	}
	if err := s.DeleteGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID); err != nil {
		t.Fatalf("DeleteGitRepository: %v", err)
	}
}

func TestStore_GitRepository_notARepository(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	dir := t.TempDir()
	_, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: dir}, gitSvc)
	if gitdomain.GitErrCode(err) != gitdomain.GitCodeNotARepository {
		t.Fatalf("got %v want not_a_git_repository", err)
	}
}

func TestStore_GitWorktreeAndBranch_roundtrip(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-a")
	wt, err := s.CreateGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-a",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktree: %v", err)
	}
	wts, err := s.ListGitWorktrees(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil || len(wts) != 2 {
		t.Fatalf("worktrees: %v len=%d", err, len(wts))
	}
	branch, err := s.CreateGitBranch(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, CreateGitBranchInput{
		Name:       "feature-b",
		StartPoint: "main",
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitBranch: %v", err)
	}
	branches, err := s.ListGitBranches(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil || len(branches) < 2 {
		t.Fatalf("branches: %v len=%d", err, len(branches))
	}
	if err := s.UnregisterGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, wt.ID); err != nil {
		t.Fatalf("UnregisterGitWorktree: %v", err)
	}
	if err := s.DeleteGitBranch(ctx, projectsdomain.LegacyGlobalDefaultProjectID, branch.ID, true, gitSvc); err != nil {
		t.Fatalf("DeleteGitBranch: %v", err)
	}
}

func TestStore_GitDeleteGuard_runningTask(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-guard")
	wt, err := s.CreateGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "guard-branch",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	branches, _ := s.ListGitBranches(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	var branchID string
	for _, b := range branches {
		if b.Name == "guard-branch" {
			branchID = b.ID
			break
		}
	}
	if branchID == "" {
		t.Fatal("guard-branch not found")
	}
	if wt.BranchID != branchID {
		t.Fatalf("worktree branch_id = %q want %q", wt.BranchID, branchID)
	}
	wtID := wt.ID
	task := model.Task{
		ID:            "task-running-guard",
		Title:         "running",
		InitialPrompt: "x",
		Status:        domain.StatusRunning,
		Priority:      domain.PriorityMedium,
		Runner:        "cursor",
		WorktreeID:    &wtID,
	}
	if err := s.db.WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatal(err)
	}
	err = s.UnregisterGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, wt.ID)
	if ge := gitdomain.GitErrCode(err); ge != gitdomain.GitCodeHasRunningTask {
		t.Fatalf("got code %q want has_running_task (%v)", ge, err)
	}
}

func TestStore_GlobalGitRepository_andWorktreeBinding(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)

	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-global")
	wt, err := s.CreateGitWorktreeForRepo(ctx, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-global",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktreeForRepo: %v", err)
	}
	if wt.BranchID == "" {
		t.Fatal("worktree missing branch_id after create")
	}
	if err := s.ValidateTaskWorktreeBinding(ctx, nil, wt.ID); err != nil {
		t.Fatalf("ValidateTaskWorktreeBinding: %v", err)
	}
	gitCtx, err := s.ResolveTaskGitContext(ctx, wt.ID)
	if err != nil {
		t.Fatalf("ResolveTaskGitContext: %v", err)
	}
	if gitCtx.WorktreePath == "" || gitCtx.BranchName != "feature-global" {
		t.Fatalf("context=%+v", gitCtx)
	}
	wt2Path := filepath.Join(filepath.Dir(main), "wt-global-2")
	_, err = s.CreateGitWorktreeForRepo(ctx, repo.ID, CreateGitWorktreeInput{
		Path:         wt2Path,
		Branch:       "feature-global",
		CreateBranch: true,
	}, gitSvc)
	if gitdomain.GitErrCode(err) != gitdomain.GitCodeBranchBoundToWorktree {
		t.Fatalf("duplicate branch on second worktree: got %v want branch_bound_to_worktree", err)
	}
}

func TestStore_ProjectRepositoryBinding(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	repoID := repo.ID
	proj, err := s.CreateProject(ctx, CreateProjectInput{
		Name:         "Overlay",
		RepositoryID: &repoID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	byRepo, err := s.ListProjectsByRepository(ctx, repo.ID)
	if err != nil || len(byRepo) != 2 {
		t.Fatalf("ListProjectsByRepository: %v len=%d want 2 (default + overlay)", err, len(byRepo))
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-proj")
	wt, err := s.CreateGitWorktreeForRepo(ctx, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "proj-branch",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	pid := proj.ID
	if err := s.ValidateTaskWorktreeBinding(ctx, &pid, wt.ID); err != nil {
		t.Fatalf("valid project binding: %v", err)
	}
	otherMain := initGitRepo(t)
	otherRepo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: otherMain}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	otherRepoID := otherRepo.ID
	otherProj, err := s.CreateProject(ctx, CreateProjectInput{
		Name:         "Other overlay",
		RepositoryID: &otherRepoID,
	})
	if err != nil {
		t.Fatal(err)
	}
	otherPID := otherProj.ID
	err = s.ValidateTaskWorktreeBinding(ctx, &otherPID, wt.ID)
	if err == nil {
		t.Fatal("expected project_repo_mismatch for project tied to different repo")
	}
	if gitdomain.GitErrCode(err) != gitdomain.GitCodeProjectRepoMismatch {
		t.Fatalf("got %v want project_repo_mismatch", err)
	}
}

func TestUnregisterGitWorktreeByID_rejectsRunningTask(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-running-global")
	wt, err := s.CreateGitWorktreeForRepo(ctx, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "running-global",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtID := wt.ID
	task := model.Task{
		ID:            "task-global-running",
		Title:         "running",
		InitialPrompt: "x",
		Status:        domain.StatusRunning,
		Priority:      domain.PriorityMedium,
		Runner:        "cursor",
		WorktreeID:    &wtID,
	}
	if err := s.db.WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatal(err)
	}
	err = s.UnregisterGitWorktreeByID(ctx, wt.ID)
	if gitdomain.GitErrCode(err) != gitdomain.GitCodeHasRunningTask {
		t.Fatalf("got %v want has_running_task", err)
	}
}
