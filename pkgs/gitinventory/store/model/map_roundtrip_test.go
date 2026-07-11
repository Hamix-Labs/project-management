package model

import (
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"reflect"
	"testing"
	"time"
)

func TestGitRepository_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	orig := gitdomain.GitRepository{
		ID: "repo-1", Path: "/repo", HostPath: "/host/repo",
		DefaultBranch: "main", CreatedAt: now, UpdatedAt: now,
	}
	m := FromDomainGitRepository(orig)
	back := ToDomainGitRepository(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}

func TestGitWorktree_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	orig := gitdomain.GitWorktree{
		ID: "wt-1", RepositoryID: "repo-1", Path: "/wt", Name: "main",
		IsMain: true, BranchID: "branch-1", CreatedAt: now,
	}
	m := FromDomainGitWorktree(orig)
	back := ToDomainGitWorktree(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}

func TestGitBranch_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	orig := gitdomain.GitBranch{
		ID: "branch-1", RepositoryID: "repo-1", Name: "main",
		HeadSHA: "abc", CreatedAt: now,
	}
	m := FromDomainGitBranch(orig)
	back := ToDomainGitBranch(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}
