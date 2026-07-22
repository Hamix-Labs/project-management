package store

import (
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
)

func TestWorktreeStaleMap_usesCycleEndedAt(t *testing.T) {
	s, ctx, _ := gitTestStore(t)
	main := initGitRepo(t)
	created, err := s.CreateGlobalGitRepository(ctx, contract.CreateGitRepositoryInput{Path: main})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}

	got, err := s.WorktreeStaleMap(ctx, created.ID, time.Now().UTC())
	if err != nil {
		t.Fatalf("WorktreeStaleMap: %v", err)
	}
	wts, err := s.ListGitWorktreesByRepo(ctx, created.ID)
	if err != nil {
		t.Fatalf("ListGitWorktreesByRepo: %v", err)
	}
	if len(wts) == 0 {
		t.Fatal("expected at least the main worktree")
	}
	for _, wt := range wts {
		if _, ok := got[wt.ID]; !ok {
			t.Fatalf("missing stale entry for worktree %s", wt.ID)
		}
		if wt.IsMain && got[wt.ID] {
			t.Fatalf("main worktree %s should not be stale", wt.ID)
		}
	}
}
