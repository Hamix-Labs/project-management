package sidecar

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAppendAndParseCommitRegister(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cycleID := "reg-1"
	if err := AppendCommitRegister(dir, cycleID, CommitRegisterEntry{
		SHA: "abc123def456", Message: "feat: one",
	}); err != nil {
		t.Fatal(err)
	}
	if err := AppendCommitRegister(dir, cycleID, CommitRegisterEntry{
		SHA: "abc123def456", Message: "dup",
	}); !errors.Is(err, ErrCommitRegisterDuplicate) {
		t.Fatalf("dup: %v", err)
	}
	if err := AppendCommitRegister(dir, cycleID, CommitRegisterEntry{
		SHA: "fed987cba654", Message: "feat: two",
	}); err != nil {
		t.Fatal(err)
	}
	got, err := ParseCommitRegister(dir, cycleID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].SHA != "abc123def456" || got[1].SHA != "fed987cba654" {
		t.Fatalf("got=%+v", got)
	}
	path := CommitRegisterPath(dir, cycleID)
	if filepath.Base(path) != commitRegisterFileName {
		t.Fatalf("path=%s", path)
	}
}

func TestParseCommitRegister_missingEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_ = EnsureReportCycleDir(dir, "c1")
	got, err := ParseCommitRegister(dir, "c1")
	if err != nil || len(got) != 0 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestParseCommitRegister_emptySHA(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cycleID := "bad"
	if err := EnsureReportCycleDir(dir, cycleID); err != nil {
		t.Fatal(err)
	}
	body := `{"schema_version":1,"commits":[{"sha":""}]}`
	if err := os.WriteFile(CommitRegisterPath(dir, cycleID), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseCommitRegister(dir, cycleID)
	if !errors.Is(err, ErrCommitRegisterInvalid) {
		t.Fatalf("got %v", err)
	}
}

func TestScrubRemovesCommitRegister(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cycleID := "scrub-reg"
	if err := AppendCommitRegister(dir, cycleID, CommitRegisterEntry{SHA: "deadbeef"}); err != nil {
		t.Fatal(err)
	}
	if err := ScrubCycleArtifacts(dir, cycleID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(CommitRegisterPath(dir, cycleID)); !os.IsNotExist(err) {
		t.Fatalf("expected register gone, err=%v", err)
	}
}
