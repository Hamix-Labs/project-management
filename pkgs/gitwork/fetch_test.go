package gitwork_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
)

func initRepoWithRemote(t *testing.T) (local, remote string) {
	t.Helper()
	requireGit(t)
	remote = t.TempDir()
	runGit(t, remote, "init", "--bare", "-b", "main")

	parent := t.TempDir()
	local = filepath.Join(parent, "local")
	runGit(t, parent, "clone", remote, local)
	runGit(t, local, "config", "user.email", "t@example.com")
	runGit(t, local, "config", "user.name", "Test")
	writeFile(t, filepath.Join(local, "README.md"), "init\n")
	runGit(t, local, "add", "README.md")
	runGit(t, local, "commit", "-m", "init")
	runGit(t, local, "push", "-u", "origin", "main")
	return local, remote
}

func TestFetch_updatesRemoteRefs(t *testing.T) {
	local, remote := initRepoWithRemote(t)
	repo := openRepo(t, local)

	otherParent := t.TempDir()
	other := filepath.Join(otherParent, "other")
	runGit(t, otherParent, "clone", remote, other)
	runGit(t, other, "config", "user.email", "t@example.com")
	runGit(t, other, "config", "user.name", "Test")
	writeFile(t, filepath.Join(other, "extra.txt"), "x\n")
	runGit(t, other, "add", "extra.txt")
	runGit(t, other, "commit", "-m", "extra")
	runGit(t, other, "push", "origin", "main")

	originBefore := runGitOutput(t, local, "rev-parse", "origin/main")
	if err := svc().Fetch(context.Background(), repo, "origin"); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	originAfter := runGitOutput(t, local, "rev-parse", "origin/main")
	if originBefore == originAfter {
		t.Fatalf("origin/main did not advance after fetch: %s", originAfter)
	}
	otherHead := runGitOutput(t, other, "rev-parse", "HEAD")
	if originAfter != otherHead {
		t.Fatalf("origin/main=%s want %s", originAfter, otherHead)
	}
}

func TestFetch_nilRepo(t *testing.T) {
	err := svc().Fetch(context.Background(), nil, "origin")
	if err != gitwork.ErrNotARepository {
		t.Fatalf("got %v want ErrNotARepository", err)
	}
}

func TestFetch_emptyRemoteDefaultsToOrigin(t *testing.T) {
	local, _ := initRepoWithRemote(t)
	repo := openRepo(t, local)
	if err := svc().Fetch(context.Background(), repo, ""); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
}

func TestResolveDefaultBranch_fromOriginHEAD(t *testing.T) {
	local, _ := initRepoWithRemote(t)
	repo := openRepo(t, local)
	if err := svc().Fetch(context.Background(), repo, ""); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	name, err := svc().ResolveDefaultBranch(context.Background(), repo, "origin")
	if err != nil {
		t.Fatalf("ResolveDefaultBranch: %v", err)
	}
	if name != "main" {
		t.Fatalf("default branch=%q want main", name)
	}
}

func TestResolveDefaultBranch_fallbackMain(t *testing.T) {
	dir := initRepo(t)
	repo := openRepo(t, dir)
	name, err := svc().ResolveDefaultBranch(context.Background(), repo, "origin")
	if err != nil {
		t.Fatalf("ResolveDefaultBranch: %v", err)
	}
	if name != "main" {
		t.Fatalf("default branch=%q want main", name)
	}
}

func TestResolveDefaultBranch_nilRepo(t *testing.T) {
	_, err := svc().ResolveDefaultBranch(context.Background(), nil, "origin")
	if err != gitwork.ErrNotARepository {
		t.Fatalf("got %v want ErrNotARepository", err)
	}
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	requireGit(t)
	all := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", all...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}
