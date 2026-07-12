package git

import (
	"context"
	"encoding/json"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/gittest"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestResetHardClean_resetsAndCleansUntracked(t *testing.T) {
	gittest.SkipIfNoGit(t)
	dir := t.TempDir()
	gittest.Init(t, dir)
	ctx := context.Background()
	repo := NewExecRepo()
	base, err := repo.Run(ctx, dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "dirty.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ResetHardClean(ctx, repo, dir, base); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("untracked file should be removed, stat err=%v", err)
	}
}

func TestResolveFreshRetryAnchor_fromExecutePhaseDetails(t *testing.T) {
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "t", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusFailed,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	exec, err := st.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}
	details, _ := json.Marshal(map[string]any{
		"git": map[string]string{"cycle_base_sha": "abc123deadbeef"},
	})
	if _, err := st.CompletePhase(ctx, cyclescontract.CompletePhaseInput{
		CycleID: cycle.ID, PhaseSeq: exec.PhaseSeq,
		Status: cyclesdomain.PhaseStatusSucceeded, Details: details, By: taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusFailed, "x", taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	svc := NewService(st, NewExecRepo(), "")
	anchor, err := svc.ResolveFreshRetryAnchor(ctx, t.TempDir(), cycle.ID)
	if err != nil {
		t.Fatal(err)
	}
	if anchor != "abc123deadbeef" {
		t.Fatalf("anchor=%q", anchor)
	}
}
