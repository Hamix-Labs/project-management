package agentmcp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/gittest"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
)

func gitConfigIdentity(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"config", "user.email", "t@e.local"},
		{"config", "user.name", "t"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func TestRunCommit_success(t *testing.T) {
	t.Parallel()
	gittest.SkipIfNoGit(t)
	dir := t.TempDir()
	gittest.Init(t, dir)
	gitConfigIdentity(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", dir, "add", "f.txt").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	reportDir := t.TempDir()
	cycleID := "c-commit"
	sess := &Session{
		TaskID:      "t1",
		CycleID:     cycleID,
		Phase:       PhaseExecute,
		ReportDir:   reportDir,
		WorkingDir:  dir,
		SubmitNonce: "n1",
	}
	out, err := runCommit(context.Background(), sess, commitInput{Message: "feat: add f"})
	if err != nil {
		t.Fatal(err)
	}
	if !out.OK || out.SHA == "" {
		t.Fatalf("out=%+v", out)
	}
	reg, err := sidecar.ParseCommitRegister(reportDir, cycleID)
	if err != nil || len(reg) != 1 || reg[0].SHA != out.SHA {
		t.Fatalf("reg=%v err=%v", reg, err)
	}
}

func TestRunCommit_nothingStaged(t *testing.T) {
	t.Parallel()
	gittest.SkipIfNoGit(t)
	dir := t.TempDir()
	gittest.Init(t, dir)
	gitConfigIdentity(t, dir)
	reportDir := t.TempDir()
	cycleID := "c-empty"
	sess := &Session{
		Phase:      PhaseExecute,
		ReportDir:  reportDir,
		WorkingDir: dir,
		CycleID:    cycleID,
	}
	_, err := runCommit(context.Background(), sess, commitInput{Message: "noop"})
	if err == nil {
		t.Fatal("expected error")
	}
	reg, err := sidecar.ParseCommitRegister(reportDir, cycleID)
	if err != nil || len(reg) != 0 {
		t.Fatalf("register must stay empty; got %v err=%v", reg, err)
	}
}

func TestRunCommit_wrongPhase(t *testing.T) {
	t.Parallel()
	sess := &Session{Phase: PhaseVerify, WorkingDir: t.TempDir()}
	_, err := runCommit(context.Background(), sess, commitInput{Message: "x"})
	if err == nil {
		t.Fatal("expected phase error")
	}
}

func TestRunCommit_emptyMessage(t *testing.T) {
	t.Parallel()
	sess := &Session{Phase: PhaseExecute, WorkingDir: t.TempDir(), ReportDir: t.TempDir(), CycleID: "c"}
	_, err := runCommit(context.Background(), sess, commitInput{Message: "  "})
	if err == nil {
		t.Fatal("expected message error")
	}
}
