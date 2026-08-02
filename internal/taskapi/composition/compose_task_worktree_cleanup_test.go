package composition_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	"github.com/google/uuid"
)

func seedRemoteMainRepo(t *testing.T) string {
	t.Helper()
	remote := t.TempDir()
	runGit(t, remote, "init", "--bare", "-b", "main")
	parent := t.TempDir()
	main := filepath.Join(parent, "main")
	runGit(t, parent, "clone", remote, main)
	runGit(t, main, "config", "user.email", "t@example.com")
	runGit(t, main, "config", "user.name", "Test")
	runGit(t, main, "commit", "--allow-empty", "-m", "init")
	runGit(t, main, "push", "-u", "origin", "main")
	return main
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
}

func TestDelete_removesManagedTaskWorktree(t *testing.T) {
	ctx := context.Background()
	t.Setenv(gitinventory.EnvManagedWorktreeRoot, t.TempDir())
	api := composition.NewAPI(tasktestdb.OpenSQLite(t))

	repo, err := api.CreateGlobalGitRepository(ctx, gitinventorystore.CreateGitRepositoryInput{Path: seedRemoteMainRepo(t)})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	taskID := uuid.NewString()
	wt, err := api.GitStore().AllocateTaskWorktree(ctx, repo.ID, taskID)
	if err != nil {
		t.Fatalf("AllocateTaskWorktree: %v", err)
	}
	wtPath := filepath.FromSlash(wt.Path)
	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("expected worktree on disk: %v", err)
	}
	branchID := wt.BranchID

	created, err := api.Create(ctx, taskcorestore.CreateTaskInput{
		ID:         taskID,
		Title:      "managed-cleanup",
		Priority:   taskcoredomain.PriorityMedium,
		WorktreeID: &wt.ID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.WorktreeID == nil || *created.WorktreeID != wt.ID {
		t.Fatalf("task worktree_id=%v want %q", created.WorktreeID, wt.ID)
	}

	if _, err := api.Delete(ctx, created.ID, taskcoredomain.ActorUser); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := api.GetGitWorktreeByID(ctx, wt.ID); err == nil {
		t.Fatal("expected worktree row removed")
	} else if gitdomain.GitErrCode(err) != gitdomain.GitCodeWorktreeNotFound {
		t.Fatalf("GetGitWorktreeByID: %v", err)
	}
	if _, err := api.GetGitBranchByID(ctx, branchID); err == nil {
		t.Fatal("expected branch row removed")
	} else if gitdomain.GitErrCode(err) != gitdomain.GitCodeBranchNotFound {
		t.Fatalf("GetGitBranchByID: %v", err)
	}
	if _, err := os.Stat(wtPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected worktree dir removed, stat=%v", err)
	}
}

func TestDelete_skipsMainWorktree(t *testing.T) {
	ctx := context.Background()
	api := composition.NewAPI(tasktestdb.OpenSQLite(t))
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", "-b", "main", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	repo, err := api.CreateGlobalGitRepository(ctx, gitinventorystore.CreateGitRepositoryInput{Path: dir})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	wts, err := api.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListGitWorktreesByRepo: %v", err)
	}
	var mainID string
	for _, wt := range wts {
		if wt.IsMain {
			mainID = wt.ID
			break
		}
	}
	if mainID == "" {
		t.Fatal("expected main worktree")
	}
	created, err := api.Create(ctx, taskcorestore.CreateTaskInput{
		Title:      "bound-to-main",
		Priority:   taskcoredomain.PriorityMedium,
		WorktreeID: &mainID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := api.Delete(ctx, created.ID, taskcoredomain.ActorUser); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := api.GetGitWorktreeByID(ctx, mainID); err != nil {
		t.Fatalf("main worktree must remain: %v", err)
	}
}

func TestDelete_skipsSharedWorktree(t *testing.T) {
	ctx := context.Background()
	t.Setenv(gitinventory.EnvManagedWorktreeRoot, t.TempDir())
	api := composition.NewAPI(tasktestdb.OpenSQLite(t))

	repo, err := api.CreateGlobalGitRepository(ctx, gitinventorystore.CreateGitRepositoryInput{Path: seedRemoteMainRepo(t)})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	ownerID := uuid.NewString()
	wt, err := api.GitStore().AllocateTaskWorktree(ctx, repo.ID, ownerID)
	if err != nil {
		t.Fatalf("AllocateTaskWorktree: %v", err)
	}
	owner, err := api.Create(ctx, taskcorestore.CreateTaskInput{
		ID:         ownerID,
		Title:      "owner",
		Priority:   taskcoredomain.PriorityMedium,
		WorktreeID: &wt.ID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("Create owner: %v", err)
	}
	shared, err := api.Create(ctx, taskcorestore.CreateTaskInput{
		Title:      "shared",
		Priority:   taskcoredomain.PriorityMedium,
		WorktreeID: &wt.ID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("Create shared: %v", err)
	}

	if _, err := api.Delete(ctx, shared.ID, taskcoredomain.ActorUser); err != nil {
		t.Fatalf("Delete shared: %v", err)
	}
	if _, err := api.GetGitWorktreeByID(ctx, wt.ID); err != nil {
		t.Fatalf("shared worktree must remain while owner exists: %v", err)
	}

	if _, err := api.Delete(ctx, owner.ID, taskcoredomain.ActorUser); err != nil {
		t.Fatalf("Delete owner: %v", err)
	}
	if _, err := api.GetGitWorktreeByID(ctx, wt.ID); err == nil {
		t.Fatal("expected worktree removed after owner delete")
	} else if gitdomain.GitErrCode(err) != gitdomain.GitCodeWorktreeNotFound {
		t.Fatalf("GetGitWorktreeByID: %v", err)
	}
}

func TestDelete_removesWorktreeWhenLastSharerDeletedAfterOwner(t *testing.T) {
	ctx := context.Background()
	t.Setenv(gitinventory.EnvManagedWorktreeRoot, t.TempDir())
	api := composition.NewAPI(tasktestdb.OpenSQLite(t))

	repo, err := api.CreateGlobalGitRepository(ctx, gitinventorystore.CreateGitRepositoryInput{Path: seedRemoteMainRepo(t)})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	ownerID := uuid.NewString()
	wt, err := api.GitStore().AllocateTaskWorktree(ctx, repo.ID, ownerID)
	if err != nil {
		t.Fatalf("AllocateTaskWorktree: %v", err)
	}
	wtPath := filepath.FromSlash(wt.Path)
	owner, err := api.Create(ctx, taskcorestore.CreateTaskInput{
		ID:         ownerID,
		Title:      "owner",
		Priority:   taskcoredomain.PriorityMedium,
		WorktreeID: &wt.ID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("Create owner: %v", err)
	}
	shared, err := api.Create(ctx, taskcorestore.CreateTaskInput{
		Title:      "shared",
		Priority:   taskcoredomain.PriorityMedium,
		WorktreeID: &wt.ID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("Create shared: %v", err)
	}

	if _, err := api.Delete(ctx, owner.ID, taskcoredomain.ActorUser); err != nil {
		t.Fatalf("Delete owner: %v", err)
	}
	if _, err := api.GetGitWorktreeByID(ctx, wt.ID); err != nil {
		t.Fatalf("worktree must remain while sharer exists: %v", err)
	}

	if _, err := api.Delete(ctx, shared.ID, taskcoredomain.ActorUser); err != nil {
		t.Fatalf("Delete shared: %v", err)
	}
	if _, err := api.GetGitWorktreeByID(ctx, wt.ID); err == nil {
		t.Fatal("expected worktree removed after last sharer delete")
	} else if gitdomain.GitErrCode(err) != gitdomain.GitCodeWorktreeNotFound {
		t.Fatalf("GetGitWorktreeByID: %v", err)
	}
	if _, err := os.Stat(wtPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected worktree dir removed, stat=%v", err)
	}
}

func TestDelete_bestEffortWhenWorktreeAlreadyGone(t *testing.T) {
	ctx := context.Background()
	t.Setenv(gitinventory.EnvManagedWorktreeRoot, t.TempDir())
	api := composition.NewAPI(tasktestdb.OpenSQLite(t))

	repo, err := api.CreateGlobalGitRepository(ctx, gitinventorystore.CreateGitRepositoryInput{Path: seedRemoteMainRepo(t)})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	taskID := uuid.NewString()
	wt, err := api.GitStore().AllocateTaskWorktree(ctx, repo.ID, taskID)
	if err != nil {
		t.Fatalf("AllocateTaskWorktree: %v", err)
	}
	created, err := api.Create(ctx, taskcorestore.CreateTaskInput{
		ID:         taskID,
		Title:      "orphan-path",
		Priority:   taskcoredomain.PriorityMedium,
		WorktreeID: &wt.ID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := api.RemoveGitWorktreeFromDiskByID(ctx, wt.ID, true); err != nil {
		t.Fatalf("RemoveGitWorktreeFromDiskByID: %v", err)
	}
	// Task still points at deleted worktree id — delete must still succeed (204 path).
	if _, err := api.Delete(ctx, created.ID, taskcoredomain.ActorUser); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
