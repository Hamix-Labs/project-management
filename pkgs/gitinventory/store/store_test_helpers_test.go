package store

import (
	"context"
	"os/exec"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
)

func gitTestStore(t *testing.T) (*Store, context.Context, gitwork.Service) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	gitSvc := gitwork.New()
	return NewStore(tasktestdb.OpenSQLite(t), gitSvc), context.Background(), gitSvc
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
