package git

import (
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
)

func TestFilterLiveWorktrees_keepsReachableDropsMissingAndPrunable(t *testing.T) {
	reachable := t.TempDir()
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	prunablePath := t.TempDir()

	live := []gitwork.Worktree{
		{Path: reachable, Branch: "feature"},
		{Path: missing, Branch: "gone"},
		{Path: prunablePath, Branch: "stale", Prunable: true},
	}

	got := FilterLiveWorktrees(live)
	if len(got) != 1 {
		t.Fatalf("len=%d want 1: %+v", len(got), got)
	}
	if got[0].Path != reachable || got[0].Branch != "feature" {
		t.Fatalf("got %+v want reachable feature worktree", got[0])
	}
}

func TestFilterLiveWorktrees_emptyInput(t *testing.T) {
	got := FilterLiveWorktrees(nil)
	if len(got) != 0 {
		t.Fatalf("len=%d want 0", len(got))
	}
}
