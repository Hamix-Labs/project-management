package sidecar

import (
	"errors"
	"os"
	"testing"
)

func TestWriteAndRequireSubmitReceipt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cycleID := "c-receipt"
	if err := WriteCriteriaReport(dir, cycleID, []CriteriaEntry{
		{ID: "a", ClaimedDone: true, Evidence: "ok"},
	}, []CriteriaCommitClaim{{SHA: "abc1234", Branch: "main"}}); err != nil {
		t.Fatal(err)
	}
	claims, err := ParseCriteriaReportCommits(dir, cycleID)
	if err != nil || len(claims) != 1 || claims[0].SHA != "abc1234" {
		t.Fatalf("commits=%v err=%v", claims, err)
	}
	path := CriteriaSubmitReceiptPath(dir, cycleID)
	if err := WriteSubmitReceipt(path, SubmitReceipt{
		Nonce: "n1", Phase: "execute", CycleID: cycleID, Tool: "hamix.submit_criteria_report",
	}); err != nil {
		t.Fatal(err)
	}
	if err := RequireCriteriaSubmitReceipt(dir, cycleID, "n1"); err != nil {
		t.Fatal(err)
	}
	if err := RequireCriteriaSubmitReceipt(dir, cycleID, "bad"); !errors.Is(err, ErrSubmitReceiptInvalid) {
		t.Fatalf("got %v", err)
	}
}

func TestRequireSubmitReceipt_missing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_ = EnsureReportCycleDir(dir, "c1")
	if err := RequireVerifySubmitReceipt(dir, "c1", "n"); !errors.Is(err, ErrSubmitReceiptMissing) {
		t.Fatalf("got %v", err)
	}
}

func TestWriteVerifyReport_roundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cycleID := "cv"
	reasoning := "Verified because the work matches the criterion and the command evidence is clean."
	if err := WriteVerifyReport(dir, cycleID, []VerifyEntry{
		{ID: "a", Verified: true, Reasoning: reasoning},
	}); err != nil {
		t.Fatal(err)
	}
	out, err := ParseVerifyReport(dir, cycleID, map[string]struct{}{"a": {}})
	if err != nil {
		t.Fatal(err)
	}
	if !out["a"].Verified {
		t.Fatalf("%+v", out["a"])
	}
	path := VerifySubmitReceiptPath(dir, cycleID)
	if err := WriteSubmitReceipt(path, SubmitReceipt{Nonce: "vn", Phase: "verify", CycleID: cycleID}); err != nil {
		t.Fatal(err)
	}
	if err := RequireVerifySubmitReceipt(dir, cycleID, "vn"); err != nil {
		t.Fatal(err)
	}
}

func TestScrubAndCleanup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cycleID := "scrub"
	if err := EnsureReportCycleDir(dir, cycleID); err != nil {
		t.Fatal(err)
	}
	marker := CriteriaReportPath(dir, cycleID)
	if err := os.WriteFile(marker, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ScrubCycleArtifacts(dir, cycleID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ReportCycleDir(dir, cycleID)); !os.IsNotExist(err) {
		t.Fatalf("expected removed dir, err=%v", err)
	}
	if err := EnsureReportCycleDir(dir, cycleID); err != nil {
		t.Fatal(err)
	}
	if err := CleanupReportDir(dir, cycleID); err != nil {
		t.Fatal(err)
	}
}

func TestParseCriteriaReportPartial_andHelpers(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cycleID := "partial"
	if err := WriteCriteriaReport(dir, cycleID, []CriteriaEntry{
		{ID: "a", ClaimedDone: false, Evidence: "wip"},
		{ID: "b", ClaimedDone: true, Evidence: "done"},
	}, nil); err != nil {
		t.Fatal(err)
	}
	partial, err := ParseCriteriaReportPartial(dir, cycleID)
	if err != nil || len(partial) != 2 {
		t.Fatalf("partial=%v err=%v", partial, err)
	}
	if MinVerifyReasoningChars() < 1 || MaxFieldBytes() < 1 {
		t.Fatal("helpers")
	}
	if CurrentSchemaVersion != 1 {
		t.Fatalf("schema %d", CurrentSchemaVersion)
	}
}
