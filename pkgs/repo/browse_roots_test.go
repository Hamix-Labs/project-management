package repo_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
)

func TestResolveWorkspacePickerRoots_expandedScope_mergesBootstrapPlaces(t *testing.T) {
	t.Parallel()
	repoPath := t.TempDir()
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatal(err)
	}
	registered := []domain.GitRepository{{
		ID:   "repo-1",
		Path: repoPath,
	}}

	defaultRoots, _, err := repo.ResolveWorkspacePickerRoots(repoPath, registered, repo.WorkspacePickerScopeDefault)
	if err != nil {
		t.Fatalf("default scope: %v", err)
	}
	if len(defaultRoots) != 1 {
		t.Fatalf("default roots len=%d want 1", len(defaultRoots))
	}

	expandedRoots, _, err := repo.ResolveWorkspacePickerRoots(repoPath, registered, repo.WorkspacePickerScopeExpanded)
	if err != nil {
		t.Fatalf("expanded scope: %v", err)
	}
	if len(expandedRoots) <= len(defaultRoots) {
		t.Fatalf("expanded roots len=%d want more than default %d", len(expandedRoots), len(defaultRoots))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	var hasHome bool
	for _, root := range expandedRoots {
		canon, err := filepath.EvalSymlinks(root.Path)
		if err != nil {
			canon = root.Path
		}
		homeCanon, err := filepath.EvalSymlinks(home)
		if err != nil {
			homeCanon = home
		}
		if filepath.Clean(canon) == filepath.Clean(homeCanon) {
			hasHome = true
		}
	}
	if !hasHome {
		t.Fatalf("expanded roots missing home: %+v", expandedRoots)
	}
}
