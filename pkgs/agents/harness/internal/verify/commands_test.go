package verify

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

func TestRunCriterionCommands_writesEvidenceAndPromptSection(t *testing.T) {
	t.Parallel()
	st := storefake.New(t).API
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         "verify-cmd",
		InitialPrompt: "do work",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
		ChecklistItems: []checklistcontract.CreateChecklistItemInput{{
			Text: "tests pass",
			VerifyCommands: []checklistcontract.VerifyCommandInput{{
				Command:         "echo hello",
				ExpectedOutcome: "prints hello",
			}},
		}},
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	items, err := st.ListChecklistForVerify(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || len(items[0].VerifyCommands) != 1 {
		t.Fatalf("verify snapshot: %+v", items)
	}

	reportDir := t.TempDir()
	workDir := t.TempDir()
	svc := NewService(Deps{
		Store:      st,
		Runner:     runnerfake.New(),
		ReportDir:  reportDir,
		WorkingDir: workDir,
	})

	cycleID := "cycle-verify-cmd"
	selfReport := map[string]reports.CriteriaEntry{
		items[0].ID: {ClaimedDone: true, Evidence: "done"},
	}
	snap := Snapshot{Criteria: items}

	evidence, err := svc.RunCriterionCommands(ctx, task.ID, cycleID, 1, 1, snap, selfReport, func(ctx context.Context, dir, command string) ([]byte, []byte, int, error) {
		if command != "echo hello" {
			t.Fatalf("command = %q", command)
		}
		return []byte("hello\n"), nil, 0, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(evidence) != 1 {
		t.Fatalf("evidence len = %d", len(evidence))
	}
	ev := evidence[0]
	for _, p := range []string{ev.StdoutPath, ev.StderrPath, ev.MetaPath} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing artifact %s: %v", p, err)
		}
	}
	stdout, _ := os.ReadFile(ev.StdoutPath)
	if string(stdout) != "hello\n" {
		t.Fatalf("stdout = %q", stdout)
	}
	metaBytes, _ := os.ReadFile(ev.MetaPath)
	var meta commandMetaFile
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		t.Fatal(err)
	}
	if meta.ExpectedOutcome != "prints hello" || meta.ExitCode != 0 {
		t.Fatalf("meta = %+v", meta)
	}
	if meta.TimeoutSeconds != nil {
		t.Fatalf("want nil TimeoutSeconds for unlimited command, got %v", meta.TimeoutSeconds)
	}

	section := FormatCommandEvidenceSection(evidence)
	if !strings.Contains(section, ev.StdoutPath) || !strings.Contains(section, "Expected outcome: prints hello") || !strings.Contains(section, "exit_code=0") {
		t.Fatalf("prompt section missing paths/details:\n%s", section)
	}

	base := filepath.Join(reportDir, cycleID, "checks", items[0].ID, "0")
	if _, err := os.Stat(base + ".stdout"); err != nil {
		t.Fatalf("expected base artifacts under %s: %v", base, err)
	}
}

func TestRunCriterionCommands_emitsLiveProgress(t *testing.T) {
	// Not parallel: mutates package-level commandProgressHeartbeat.
	prevHeartbeat := commandProgressHeartbeat
	commandProgressHeartbeat = 20 * time.Millisecond
	t.Cleanup(func() { commandProgressHeartbeat = prevHeartbeat })

	st := storefake.New(t).API
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         "verify-cmd-progress",
		InitialPrompt: "do work",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
		ChecklistItems: []checklistcontract.CreateChecklistItemInput{{
			Text: "tests pass",
			VerifyCommands: []checklistcontract.VerifyCommandInput{{
				Command:         "go test ./...",
				ExpectedOutcome: "Exit code 0.",
			}},
		}},
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	items, err := st.ListChecklistForVerify(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var events []runner.ProgressEvent
	svc := NewService(Deps{
		Store:      st,
		Runner:     runnerfake.New(),
		ReportDir:  t.TempDir(),
		WorkingDir: t.TempDir(),
		Hooks: Hooks{
			PersistProgress: func(ctx context.Context, taskID, cycleID string, phaseSeq int64, ev runner.ProgressEvent) {
				if taskID != task.ID || cycleID != "cycle-progress" || phaseSeq != 2 {
					t.Errorf("unexpected progress target task=%s cycle=%s phase=%d", taskID, cycleID, phaseSeq)
				}
				mu.Lock()
				events = append(events, ev)
				mu.Unlock()
			},
		},
	})

	selfReport := map[string]reports.CriteriaEntry{
		items[0].ID: {ClaimedDone: true, Evidence: "done"},
	}
	snap := Snapshot{Criteria: items}

	_, err = svc.RunCriterionCommands(ctx, task.ID, "cycle-progress", 2, 1, snap, selfReport,
		func(ctx context.Context, dir, command string) ([]byte, []byte, int, error) {
			time.Sleep(55 * time.Millisecond)
			return []byte("ok\n"), nil, 0, nil
		})
	if err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(events) < 3 {
		t.Fatalf("want at least started+running+completed, got %d: %+v", len(events), events)
	}
	if events[0].Subtype != "started" || events[0].Tool != ProgressToolVerifyCommand {
		t.Fatalf("first event = %+v", events[0])
	}
	if !strings.Contains(events[0].Message, "Running: go test") {
		t.Fatalf("started message = %q", events[0].Message)
	}
	sawRunning := false
	for _, ev := range events[1 : len(events)-1] {
		if ev.Subtype == "running" {
			sawRunning = true
			if !strings.Contains(ev.Message, "Running: go test") {
				t.Fatalf("running message = %q", ev.Message)
			}
		}
	}
	if !sawRunning {
		t.Fatalf("missing running heartbeat in %+v", events)
	}
	last := events[len(events)-1]
	if last.Subtype != "completed" || !strings.Contains(last.Message, "Finished:") {
		t.Fatalf("last event = %+v", last)
	}
}

func TestRunCriterionCommands_heartbeatsContinuePastExecTimeout(t *testing.T) {
	// Not parallel: mutates package-level commandProgressHeartbeat.
	prevHeartbeat := commandProgressHeartbeat
	commandProgressHeartbeat = 25 * time.Millisecond
	t.Cleanup(func() { commandProgressHeartbeat = prevHeartbeat })

	st := storefake.New(t).API
	ctx := context.Background()

	execTimeoutSec := 1
	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         "verify-cmd-past-timeout",
		InitialPrompt: "do work",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
		ChecklistItems: []checklistcontract.CreateChecklistItemInput{{
			Text: "tests pass",
			VerifyCommands: []checklistcontract.VerifyCommandInput{{
				Command:         "slow-check",
				ExpectedOutcome: "Exit code 0.",
				TimeoutSeconds:  &execTimeoutSec,
			}},
		}},
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	items, err := st.ListChecklistForVerify(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var events []runner.ProgressEvent
	svc := NewService(Deps{
		Store:      st,
		Runner:     runnerfake.New(),
		ReportDir:  t.TempDir(),
		WorkingDir: t.TempDir(),
		Hooks: Hooks{
			PersistProgress: func(ctx context.Context, taskID, cycleID string, phaseSeq int64, ev runner.ProgressEvent) {
				mu.Lock()
				events = append(events, ev)
				mu.Unlock()
			},
		},
	})

	selfReport := map[string]reports.CriteriaEntry{
		items[0].ID: {ClaimedDone: true, Evidence: "done"},
	}
	snap := Snapshot{Criteria: items}

	timeoutFired := make(chan struct{})
	_, err = svc.RunCriterionCommands(ctx, task.ID, "cycle-past-timeout", 2, 1, snap, selfReport,
		func(cmdCtx context.Context, dir, command string) ([]byte, []byte, int, error) {
			select {
			case <-cmdCtx.Done():
				close(timeoutFired)
			case <-time.After(2 * time.Second):
				t.Fatal("execCtx did not cancel within 2s")
			}
			// Keep Wait outstanding past the kill timer so progress must outlive execCtx.
			time.Sleep(80 * time.Millisecond)
			return []byte("late\n"), nil, 1, nil
		})
	if err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(events) < 3 {
		t.Fatalf("want started + timeout/running + failed, got %d: %+v", len(events), events)
	}
	if events[0].Subtype != "started" {
		t.Fatalf("first event = %+v", events[0])
	}
	sawTimeoutBoundary := false
	sawPostTimeoutHeartbeat := false
	for _, ev := range events[1 : len(events)-1] {
		if ev.Subtype != "running" {
			continue
		}
		if strings.Contains(ev.Message, "Timed out waiting for:") {
			sawTimeoutBoundary = true
			continue
		}
		if sawTimeoutBoundary && strings.Contains(ev.Message, "Running: slow-check") {
			sawPostTimeoutHeartbeat = true
		}
	}
	if !sawTimeoutBoundary {
		t.Fatalf("missing timeout-boundary progress in %+v", events)
	}
	if !sawPostTimeoutHeartbeat {
		t.Fatalf("missing heartbeat after exec timeout in %+v", events)
	}
	last := events[len(events)-1]
	if last.Subtype != "failed" || !strings.Contains(last.Message, "Failed:") {
		t.Fatalf("last event = %+v", last)
	}
	select {
	case <-timeoutFired:
	default:
		t.Fatal("execFn never observed execCtx cancellation")
	}
}

func TestRunCriterionCommands_noTimeoutUsesParentCtxOnly(t *testing.T) {
	t.Parallel()
	st := storefake.New(t).API
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         "verify-cmd-no-timeout",
		InitialPrompt: "do work",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
		ChecklistItems: []checklistcontract.CreateChecklistItemInput{{
			Text: "tests pass",
			VerifyCommands: []checklistcontract.VerifyCommandInput{{
				Command:         "slow-unlimited",
				ExpectedOutcome: "Exit code 0.",
			}},
		}},
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	items, err := st.ListChecklistForVerify(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}

	svc := NewService(Deps{
		Store:      st,
		Runner:     runnerfake.New(),
		ReportDir:  t.TempDir(),
		WorkingDir: t.TempDir(),
	})
	selfReport := map[string]reports.CriteriaEntry{
		items[0].ID: {ClaimedDone: true, Evidence: "done"},
	}
	snap := Snapshot{Criteria: items}

	sawDeadline := false
	evidence, err := svc.RunCriterionCommands(ctx, task.ID, "cycle-no-timeout", 1, 1, snap, selfReport,
		func(cmdCtx context.Context, dir, command string) ([]byte, []byte, int, error) {
			if _, ok := cmdCtx.Deadline(); ok {
				sawDeadline = true
			}
			// Would have been killed under the old 120s global default if a short
			// timeout were wrongly applied; here we only sleep briefly.
			time.Sleep(50 * time.Millisecond)
			select {
			case <-cmdCtx.Done():
				t.Fatal("execCtx cancelled despite no timeout_seconds")
			default:
			}
			return []byte("ok\n"), nil, 0, nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if sawDeadline {
		t.Fatal("execCtx unexpectedly had a deadline when timeout_seconds was omitted")
	}
	if len(evidence) != 1 || evidence[0].ExitCode != 0 {
		t.Fatalf("evidence = %+v", evidence)
	}
}

func TestProgressStreamHelpers(t *testing.T) {
	t.Parallel()
	got := truncateProgressCommand(strings.Repeat("a", 200))
	if got != strings.Repeat("a", maxProgressCommandChars-1)+"…" {
		t.Fatalf("truncate = %q (rune len %d)", got, len([]rune(got)))
	}
	if got := formatProgressElapsed(5 * time.Second); got != "5s" {
		t.Fatalf("elapsed = %q", got)
	}
	if got := formatProgressElapsed(125 * time.Second); got != "2m 5s" {
		t.Fatalf("elapsed = %q", got)
	}
}
