package store

import (
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

func TestStore_UnregisterGitWorktreeByID_preservesDiskCheckout(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)

	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-persist")
	wt, err := s.CreateGitWorktreeForRepo(ctx, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-persist",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktreeForRepo: %v", err)
	}
	if err := s.UnregisterGitWorktreeByID(ctx, wt.ID); err != nil {
		t.Fatalf("UnregisterGitWorktreeByID: %v", err)
	}
	wts, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListGitWorktreesByRepo: %v", err)
	}
	if len(wts) != 1 {
		t.Fatalf("registered worktrees len=%d want 1 (main only)", len(wts))
	}
	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("checkout path should remain on disk after unregister: %v", err)
	}
	inventory, err := s.RepoWorktreeInventory(ctx, repo, gitSvc)
	if err != nil {
		t.Fatalf("RepoWorktreeInventory: %v", err)
	}
	invRow, ok := FindWorktreeInInventory(inventory, wtPath)
	if !ok {
		t.Fatalf("linked worktree missing from live inventory after unregister")
	}
	if invRow.Registered {
		t.Fatal("worktree should be unregistered in live inventory after unregister")
	}
}

func TestStore_RemoveGitWorktreeFromDiskByID_removesCheckout(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)

	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-remove-disk")
	wt, err := s.CreateGitWorktreeForRepo(ctx, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-remove-disk",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktreeForRepo: %v", err)
	}
	if err := s.RemoveGitWorktreeFromDiskByID(ctx, wt.ID, false, gitSvc); err != nil {
		t.Fatalf("RemoveGitWorktreeFromDiskByID: %v", err)
	}
	wts, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListGitWorktreesByRepo: %v", err)
	}
	if len(wts) != 1 {
		t.Fatalf("registered worktrees len=%d want 1 (main only)", len(wts))
	}
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Fatalf("checkout path should be removed from disk, stat err=%v", err)
	}
}

func TestStore_RemoveGitWorktreeFromDiskByID_rejectsMain(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)

	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	wts, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	var mainWt gitdomain.GitWorktree
	for _, wt := range wts {
		if wt.IsMain {
			mainWt = wt
			break
		}
	}
	if mainWt.ID == "" {
		t.Fatal("main worktree not found")
	}
	err = s.RemoveGitWorktreeFromDiskByID(ctx, mainWt.ID, false, gitSvc)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v want ErrInvalidInput", err)
	}
}

func TestRelocateGitWorktree_updatesRegisteredPath(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-reloc-store")
	wt, err := s.CreateGitWorktreeForRepo(ctx, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "reloc-store",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	movedPath := filepath.Join(filepath.Dir(main), "wt-reloc-store-moved")
	runGitStore(t, main, "worktree", "move", wtPath, movedPath)
	t.Cleanup(func() { _ = os.RemoveAll(movedPath) })

	got, err := s.RelocateGitWorktree(ctx, wt.ID, movedPath, gitSvc)
	if err != nil {
		t.Fatalf("RelocateGitWorktree: %v", err)
	}
	if worktreePathKey(got.Path) != worktreePathKey(movedPath) {
		t.Fatalf("path=%q want %q", got.Path, movedPath)
	}
}

func TestCreateGitBranchForRepo_global(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	br, err := s.CreateGitBranchForRepo(ctx, repo.ID, CreateGitBranchInput{
		Name:       "feature-global-branch",
		StartPoint: "main",
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitBranchForRepo: %v", err)
	}
	if br.Name != "feature-global-branch" || br.HeadSHA == "" {
		t.Fatalf("branch=%+v", br)
	}
}
