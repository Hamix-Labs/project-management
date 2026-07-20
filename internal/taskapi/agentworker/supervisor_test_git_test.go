package agentworker_test

import (
	"context"
	"os/exec"
	"testing"

	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
)

func (rig *supervisorTestRig) seedGitRepository(t *testing.T, dir string) {
	t.Helper()
	if dir == "" {
		dir = t.TempDir()
	}
	if out, err := exec.Command("git", "init", "-b", "main", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	for _, args := range [][]string{
		{"config", "user.email", "sup@test.local"},
		{"config", "user.name", "Sup Test"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	_ = exec.Command("git", "-C", dir, "commit", "-m", "init", "--allow-empty").Run()
	ctx := context.Background()
	if _, err := rig.store.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, gitinventorystore.CreateGitRepositoryInput{
		Path: dir,
	}); err != nil {
		t.Fatalf("CreateGitRepository: %v", err)
	}
}

func (rig *supervisorTestRig) seedRunnableWorker(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	rig.seedGitRepository(t, dir)
	return dir
}
