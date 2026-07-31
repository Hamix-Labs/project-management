package runnerfake

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// SeedCommitRegisterForTests creates an allow-empty commit and appends it to the
// cycle commit register so worker/e2e fakes satisfy ADR-0093 ingest without
// scripting hamix.commit. No-op when not an execute visit or paths are incomplete.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only commit register seed."
func SeedCommitRegisterForTests(req runner.Request) error {
	return seedCommitRegisterForTests(req)
}

//funclogmeasure:skip category=tool-required-noop reason="Test-only commit register seed."
func seedCommitRegisterForTests(req runner.Request) error {
	if req.Phase != cyclesdomain.PhaseExecute {
		return nil
	}
	workDir := strings.TrimSpace(req.WorkingDir)
	reportDir := strings.TrimSpace(req.ReportDir)
	cycleID := strings.TrimSpace(req.CycleID)
	if workDir == "" || reportDir == "" || cycleID == "" {
		return nil
	}
	if err := exec.Command("git", "-C", workDir, "rev-parse", "--git-dir").Run(); err != nil {
		return nil // non-git workdir — harness skips ingest
	}
	existing, err := sidecar.ParseCommitRegister(reportDir, cycleID)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}
	msgFile, err := os.CreateTemp("", "runnerfake-commit-*.txt")
	if err != nil {
		return err
	}
	msgPath := msgFile.Name()
	defer os.Remove(msgPath)
	if _, err := msgFile.WriteString("test: seed commit register\n"); err != nil {
		_ = msgFile.Close()
		return err
	}
	if err := msgFile.Close(); err != nil {
		return err
	}
	cmd := exec.Command("git", "-C", workDir,
		"-c", "user.email=fake@test.local",
		"-c", "user.name=runnerfake",
		"commit", "--allow-empty", "-F", msgPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit --allow-empty: %w\n%s", err, out)
	}
	shaOut, err := exec.Command("git", "-C", workDir, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("rev-parse HEAD: %w\n%s", err, shaOut)
	}
	sha := strings.TrimSpace(string(shaOut))
	return sidecar.AppendCommitRegister(reportDir, cycleID, sidecar.CommitRegisterEntry{
		SHA:     sha,
		Message: "test: seed commit register",
	})
}
