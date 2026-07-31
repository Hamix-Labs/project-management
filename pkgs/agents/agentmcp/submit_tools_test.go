package agentmcp

import (
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
)

func TestSubmitCriteria_writesReportAndReceipt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cycleID := "cycle-1"
	nonce := "nonce-abc"
	sess := &Session{
		TaskID:             "task-1",
		CycleID:            cycleID,
		Phase:              PhaseExecute,
		ReportDir:          dir,
		SubmitNonce:        nonce,
		ActiveCriterionIDs: map[string]struct{}{"a": {}},
	}
	out, err := submitCriteria(sess, submitCriteriaInput{
		Criteria: []struct {
			ID          string `json:"id" jsonschema:"criterion id"`
			ClaimedDone bool   `json:"claimed_done" jsonschema:"whether the criterion is claimed done"`
			Evidence    string `json:"evidence" jsonschema:"evidence for the claim"`
		}{
			{ID: "a", ClaimedDone: true, Evidence: "done"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Fatalf("out=%+v", out)
	}
	got, err := sidecar.ParseCriteriaReport(dir, cycleID, map[string]struct{}{"a": {}})
	if err != nil {
		t.Fatal(err)
	}
	if !got["a"].ClaimedDone {
		t.Fatalf("entry=%+v", got["a"])
	}
	if err := sidecar.RequireCriteriaSubmitReceipt(dir, cycleID, nonce); err != nil {
		t.Fatal(err)
	}
}

func TestSubmitCriteria_wrongPhase(t *testing.T) {
	t.Parallel()
	sess := &Session{Phase: PhaseVerify, ActiveCriterionIDs: map[string]struct{}{"a": {}}}
	_, err := submitCriteria(sess, submitCriteriaInput{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSubmitCriteria_missingActiveID(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sess := &Session{
		Phase:              PhaseExecute,
		ReportDir:          dir,
		CycleID:            "c1",
		SubmitNonce:        "n",
		ActiveCriterionIDs: map[string]struct{}{"a": {}, "b": {}},
	}
	_, err := submitCriteria(sess, submitCriteriaInput{
		Criteria: []struct {
			ID          string `json:"id" jsonschema:"criterion id"`
			ClaimedDone bool   `json:"claimed_done" jsonschema:"whether the criterion is claimed done"`
			Evidence    string `json:"evidence" jsonschema:"evidence for the claim"`
		}{
			{ID: "a", ClaimedDone: true, Evidence: "x"},
		},
	})
	if err == nil {
		t.Fatal("expected missing b")
	}
}

func TestLoadBind_roundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "bind.json")
	b := BindFile{
		TaskID:             "t1",
		CycleID:            "c1",
		Phase:              PhaseExecute,
		ReportDir:          dir,
		WorkingDir:         dir,
		ActiveCriterionIDs: []string{"a"},
		SubmitNonce:        "n1",
	}
	if err := WriteBind(path, b); err != nil {
		t.Fatal(err)
	}
	sess, err := LoadBind(path)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Phase != PhaseExecute || sess.SubmitNonce != "n1" {
		t.Fatalf("%+v", sess)
	}
	if _, ok := sess.ActiveCriterionIDs["a"]; !ok {
		t.Fatal("missing active a")
	}
}
