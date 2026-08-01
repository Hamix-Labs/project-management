package git

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/gittest"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func TestIngestExecuteCommits_fromRegister(t *testing.T) {
	t.Parallel()
	gittest.SkipIfNoGit(t)
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "t", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	gittest.Init(t, dir)
	repo := NewExecRepo()
	base, err := repo.Run(ctx, dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "f.go"), []byte("package p\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "f.go"}, {"commit", "-m", "feat"}} {
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.email=t@e.local", "-c", "user.name=t"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	head, err := repo.Run(ctx, dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	head = trimLine(head)

	reportDir := t.TempDir()
	if err := sidecar.AppendCommitRegister(reportDir, cycle.ID, sidecar.CommitRegisterEntry{SHA: head, Message: "feat"}); err != nil {
		t.Fatal(err)
	}

	svc := NewService(st, repo, reportDir)
	snap := PhaseSnapshot{
		Repo:         dir,
		Worktree:     dir,
		BaseSHA:      trimLine(base),
		CycleBaseSHA: trimLine(base),
	}
	outcome, err := svc.IngestExecuteCommits(ctx, tsk.ID, cycle, 1, snap, nil, IngestExecuteCommitsOpts{})
	if err != nil {
		t.Fatalf("ingest err: %v", err)
	}
	if outcome.FailReason != "" {
		t.Fatalf("want no fail reason, got %q", outcome.FailReason)
	}
	if outcome.CommitCount != 1 {
		t.Fatalf("commit_count=%d want 1", outcome.CommitCount)
	}
	rows, err := st.ListCommitsForTask(ctx, tsk.ID)
	if err != nil || len(rows) != 1 {
		t.Fatalf("task commits: %v len=%d", err, len(rows))
	}
	if rows[0].SHA != head {
		t.Fatalf("sha=%s want %s", rows[0].SHA, head)
	}
}

func TestIngestExecuteCommits_emptyRegisterFails(t *testing.T) {
	t.Parallel()
	gittest.SkipIfNoGit(t)
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "t", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	gittest.Init(t, dir)
	repo := NewExecRepo()
	base, _ := repo.Run(ctx, dir, "rev-parse", "HEAD")
	svc := NewService(st, repo, t.TempDir())
	snap := PhaseSnapshot{Repo: dir, Worktree: dir, BaseSHA: trimLine(base), CycleBaseSHA: trimLine(base)}
	outcome, err := svc.IngestExecuteCommits(ctx, tsk.ID, cycle, 1, snap, nil, IngestExecuteCommitsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.FailReason != ExecuteMissingCommitsReason {
		t.Fatalf("got %+v", outcome)
	}
}

func TestIngestExecuteCommits_emptyRegisterAllowedWhenNoNewCommits(t *testing.T) {
	t.Parallel()
	gittest.SkipIfNoGit(t)
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "t", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	gittest.Init(t, dir)
	repo := NewExecRepo()
	base, _ := repo.Run(ctx, dir, "rev-parse", "HEAD")
	svc := NewService(st, repo, t.TempDir())
	snap := PhaseSnapshot{Repo: dir, Worktree: dir, BaseSHA: trimLine(base), CycleBaseSHA: trimLine(base)}
	outcome, err := svc.IngestExecuteCommits(ctx, tsk.ID, cycle, 1, snap, nil, IngestExecuteCommitsOpts{AllowEmptyRegister: true})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.FailReason != "" || outcome.CommitCount != 0 {
		t.Fatalf("got %+v", outcome)
	}
}

func TestIngestExecuteCommits_shellOnlyUnregistered(t *testing.T) {
	t.Parallel()
	gittest.SkipIfNoGit(t)
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "t", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	gittest.Init(t, dir)
	repo := NewExecRepo()
	base, _ := repo.Run(ctx, dir, "rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(dir, "x.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "x.txt"}, {"commit", "-m", "shell"}} {
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.email=t@e.local", "-c", "user.name=t"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	svc := NewService(st, repo, t.TempDir())
	snap := PhaseSnapshot{Repo: dir, Worktree: dir, BaseSHA: trimLine(base), CycleBaseSHA: trimLine(base)}
	outcome, err := svc.IngestExecuteCommits(ctx, tsk.ID, cycle, 1, snap, nil, IngestExecuteCommitsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	// Empty register → missing takes precedence over unregistered when R is empty.
	if outcome.FailReason != ExecuteMissingCommitsReason {
		t.Fatalf("got %+v", outcome)
	}
}

func TestIngestExecuteCommits_partialRegisterUnregistered(t *testing.T) {
	t.Parallel()
	gittest.SkipIfNoGit(t)
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "t", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	gittest.Init(t, dir)
	repo := NewExecRepo()
	base, _ := repo.Run(ctx, dir, "rev-parse", "HEAD")
	reportDir := t.TempDir()

	commitFile := func(name, msg string) string {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"add", name}, {"commit", "-m", msg}} {
			cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.email=t@e.local", "-c", "user.name=t"}, args...)...)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v\n%s", args, err, out)
			}
		}
		h, err := repo.Run(ctx, dir, "rev-parse", "HEAD")
		if err != nil {
			t.Fatal(err)
		}
		return trimLine(h)
	}
	first := commitFile("a.txt", "one")
	_ = commitFile("b.txt", "two") // shell-only second commit
	if err := sidecar.AppendCommitRegister(reportDir, cycle.ID, sidecar.CommitRegisterEntry{SHA: first}); err != nil {
		t.Fatal(err)
	}
	svc := NewService(st, repo, reportDir)
	snap := PhaseSnapshot{Repo: dir, Worktree: dir, BaseSHA: trimLine(base), CycleBaseSHA: trimLine(base)}
	outcome, err := svc.IngestExecuteCommits(ctx, tsk.ID, cycle, 1, snap, nil, IngestExecuteCommitsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.FailReason != ExecuteUnregisteredCommitsReason {
		t.Fatalf("got %+v", outcome)
	}
}

func TestIngestExecuteCommits_fabricatedRegisterSHA(t *testing.T) {
	t.Parallel()
	gittest.SkipIfNoGit(t)
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "t", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	gittest.Init(t, dir)
	repo := NewExecRepo()
	base, _ := repo.Run(ctx, dir, "rev-parse", "HEAD")
	reportDir := t.TempDir()
	if err := sidecar.AppendCommitRegister(reportDir, cycle.ID, sidecar.CommitRegisterEntry{
		SHA: "ffffffffffffffffffffffffffffffffffffffff",
	}); err != nil {
		t.Fatal(err)
	}
	svc := NewService(st, repo, reportDir)
	snap := PhaseSnapshot{Repo: dir, Worktree: dir, BaseSHA: trimLine(base), CycleBaseSHA: trimLine(base)}
	outcome, err := svc.IngestExecuteCommits(ctx, tsk.ID, cycle, 1, snap, nil, IngestExecuteCommitsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.FailReason != ExecuteInvalidCommitReason {
		t.Fatalf("got %+v", outcome)
	}
}

func TestIngestExecuteCommits_twoCommits(t *testing.T) {
	t.Parallel()
	gittest.SkipIfNoGit(t)
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "t", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	gittest.Init(t, dir)
	repo := NewExecRepo()
	base, _ := repo.Run(ctx, dir, "rev-parse", "HEAD")
	reportDir := t.TempDir()
	var shas []string
	for i, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"add", name}, {"commit", "-m", name}} {
			cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.email=t@e.local", "-c", "user.name=t"}, args...)...)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v\n%s", args, err, out)
			}
		}
		h, _ := repo.Run(ctx, dir, "rev-parse", "HEAD")
		h = trimLine(h)
		shas = append(shas, h)
		if err := sidecar.AppendCommitRegister(reportDir, cycle.ID, sidecar.CommitRegisterEntry{SHA: h, Message: name}); err != nil {
			t.Fatal(err)
		}
		_ = i
	}
	svc := NewService(st, repo, reportDir)
	snap := PhaseSnapshot{Repo: dir, Worktree: dir, BaseSHA: trimLine(base), CycleBaseSHA: trimLine(base)}
	outcome, err := svc.IngestExecuteCommits(ctx, tsk.ID, cycle, 1, snap, nil, IngestExecuteCommitsOpts{})
	if err != nil || outcome.FailReason != "" || outcome.CommitCount != 2 {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
	rows, err := st.ListCommitsForTask(ctx, tsk.ID)
	if err != nil || len(rows) != 2 {
		t.Fatalf("rows=%d err=%v", len(rows), err)
	}
	if rows[0].SHA != shas[0] || rows[1].SHA != shas[1] {
		t.Fatalf("order got %s,%s want %s,%s", rows[0].SHA, rows[1].SHA, shas[0], shas[1])
	}
}

func TestIngestExecuteCommits_skippedNoop(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "t", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(st, NewExecRepo(), t.TempDir())
	outcome, err := svc.IngestExecuteCommits(ctx, tsk.ID, cycle, 1, PhaseSnapshot{Skipped: true}, nil, IngestExecuteCommitsOpts{})
	if err != nil || outcome.FailReason != "" || outcome.CommitCount != 0 {
		t.Fatalf("got %+v err=%v", outcome, err)
	}
}

func TestIngestExecuteCommits_legacyCriteriaCommitsIgnored(t *testing.T) {
	t.Parallel()
	gittest.SkipIfNoGit(t)
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "t", InitialPrompt: "p", Priority: taskcoredomain.PriorityMedium, Status: taskcoredomain.StatusReady,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	gittest.Init(t, dir)
	repo := NewExecRepo()
	base, _ := repo.Run(ctx, dir, "rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(dir, "f.go"), []byte("package p\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "f.go"}, {"commit", "-m", "feat"}} {
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.email=t@e.local", "-c", "user.name=t"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	head, _ := repo.Run(ctx, dir, "rev-parse", "HEAD")
	head = trimLine(head)
	reportDir := t.TempDir()
	cdir := filepath.Join(reportDir, cycle.ID)
	if err := os.MkdirAll(cdir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Legacy criteria-report commits[] alone must not satisfy ingest.
	raw, _ := json.Marshal(map[string]any{
		"schema_version": 1,
		"criteria":       []map[string]any{{"id": "c1", "claimed_done": true, "evidence": "done"}},
		"commits":        []map[string]any{{"sha": head}},
	})
	if err := os.WriteFile(filepath.Join(cdir, "criteria-report.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	svc := NewService(st, repo, reportDir)
	snap := PhaseSnapshot{Repo: dir, Worktree: dir, BaseSHA: trimLine(base), CycleBaseSHA: trimLine(base)}
	outcome, err := svc.IngestExecuteCommits(ctx, tsk.ID, cycle, 1, snap, nil, IngestExecuteCommitsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.FailReason != ExecuteMissingCommitsReason {
		t.Fatalf("legacy commits[] must be ignored; got %+v", outcome)
	}
	// Register alone drives success.
	if err := sidecar.AppendCommitRegister(reportDir, cycle.ID, sidecar.CommitRegisterEntry{SHA: head}); err != nil {
		t.Fatal(err)
	}
	outcome, err = svc.IngestExecuteCommits(ctx, tsk.ID, cycle, 1, snap, nil, IngestExecuteCommitsOpts{})
	if err != nil || outcome.FailReason != "" || outcome.CommitCount != 1 {
		t.Fatalf("got %+v err=%v", outcome, err)
	}
}

func trimLine(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

func TestCriteriaReportProbeErr_roundTrip(t *testing.T) {
	t.Parallel()
	base, _ := json.Marshal(map[string]any{"summary": "ok"})
	merged := MergeCriteriaReportProbeErr(base, "criteria report invalid: unknown field")
	got := CriteriaReportProbeErrFromPhaseDetails(merged)
	if got != "criteria report invalid: unknown field" {
		t.Fatalf("got %q", got)
	}
}
