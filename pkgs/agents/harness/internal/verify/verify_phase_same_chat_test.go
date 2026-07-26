package verify_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// TestWorker_VerifyPhase_resumesExecuteSession pins ADR-0085: first verify
// after execute --resumes the execute phase session_id (same Cursor chat).
func TestWorker_VerifyPhase_resumesExecuteSession(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	enabled := true
	if _, err := h.Store.UpdateSettings(ctx, settingscontract.SettingsPatch{
		CursorSessionResumeEnabled: &enabled,
	}); err != nil {
		t.Fatalf("enable cursor resume: %v", err)
	}

	tsk := h.CreateReadyTask(ctx, "verify-same-chat")
	item, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion one", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add checklist item: %v", err)
	}

	workDir := t.TempDir()
	reportDir := t.TempDir()
	const execSession = "sess-execute-1"

	execDetails, err := json.Marshal(map[string]string{"session_id": execSession})
	if err != nil {
		t.Fatal(err)
	}

	r := runnerfake.New()
	hook := &hookRunner{Runner: r, preRun: func(req runner.Request) {
		cycles, _ := h.Store.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		switch req.Phase {
		case cyclesdomain.PhaseExecute:
			writeCriteriaReport(t, reportDir, cycles[0].ID, []string{item.ID})
		case cyclesdomain.PhaseVerify:
			writeVerifyReport(t, reportDir, cycles[0].ID, []string{item.ID})
		}
	}}
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "exec ok", execDetails, ""))
	r.Script(tsk.ID, cyclesdomain.PhaseVerify, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "verify ok", execDetails, ""))

	done := h.StartHarnessRun(ctx, tsk, hook, harness.Options{
		WorkingDir: workDir,
		ReportDir:  reportDir,
	})
	h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
	<-done
	cancel()

	var verifyReq *runner.Request
	for i := range r.Calls() {
		c := r.Calls()[i]
		if c.Phase == cyclesdomain.PhaseVerify {
			verifyReq = &c
			break
		}
	}
	if verifyReq == nil {
		t.Fatal("expected a verify runner call")
	}
	if verifyReq.ResumeSessionID != execSession {
		t.Fatalf("verify ResumeSessionID = %q, want execute session %q", verifyReq.ResumeSessionID, execSession)
	}
	if !strings.Contains(verifyReq.Prompt, "verify-report.json") {
		t.Fatalf("same-chat verify prompt missing verify-report path: %q", verifyReq.Prompt)
	}
	if !strings.Contains(verifyReq.Prompt, `Schema: {"criteria"`) {
		t.Fatalf("same-chat verify prompt missing schema: %q", verifyReq.Prompt)
	}
	if strings.Contains(verifyReq.Prompt, "criteria-report.json") {
		t.Fatalf("same-chat verify prompt must not ask for criteria-report.json: %q", verifyReq.Prompt)
	}
}

func TestWorker_VerifyPhase_passesVerifyModelWithSameSession(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	enabled := true
	verifyModel := "composer-2.5-fast"
	if _, err := h.Store.UpdateSettings(ctx, settingscontract.SettingsPatch{
		CursorSessionResumeEnabled: &enabled,
		VerifyModel:                &verifyModel,
	}); err != nil {
		t.Fatalf("settings: %v", err)
	}

	tsk := h.CreateReadyTask(ctx, "verify-model-same-chat")
	item, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion one", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add checklist item: %v", err)
	}

	workDir := t.TempDir()
	reportDir := t.TempDir()
	const execSession = "sess-execute-model"

	execDetails, err := json.Marshal(map[string]string{"session_id": execSession})
	if err != nil {
		t.Fatal(err)
	}

	r := runnerfake.New()
	hook := &hookRunner{Runner: r, preRun: func(req runner.Request) {
		cycles, _ := h.Store.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		switch req.Phase {
		case cyclesdomain.PhaseExecute:
			writeCriteriaReport(t, reportDir, cycles[0].ID, []string{item.ID})
		case cyclesdomain.PhaseVerify:
			writeVerifyReport(t, reportDir, cycles[0].ID, []string{item.ID})
		}
	}}
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "exec ok", execDetails, ""))
	r.Script(tsk.ID, cyclesdomain.PhaseVerify, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "verify ok", execDetails, ""))

	done := h.StartHarnessRun(ctx, tsk, hook, harness.Options{
		WorkingDir: workDir,
		ReportDir:  reportDir,
	})
	h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusReview)
	<-done
	cancel()

	var verifyReq *runner.Request
	for i := range r.Calls() {
		c := r.Calls()[i]
		if c.Phase == cyclesdomain.PhaseVerify {
			verifyReq = &c
			break
		}
	}
	if verifyReq == nil {
		t.Fatal("expected a verify runner call")
	}
	if verifyReq.ResumeSessionID != execSession {
		t.Fatalf("ResumeSessionID = %q, want %q", verifyReq.ResumeSessionID, execSession)
	}
	if verifyReq.CursorModel != verifyModel {
		t.Fatalf("CursorModel = %q, want %q", verifyReq.CursorModel, verifyModel)
	}
}

func TestWorker_VerifyPhase_missingExecuteSessionHardFailsWithoutRun(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	enabled := true
	if _, err := h.Store.UpdateSettings(ctx, settingscontract.SettingsPatch{
		CursorSessionResumeEnabled: &enabled,
	}); err != nil {
		t.Fatalf("enable cursor resume: %v", err)
	}

	tsk := h.CreateReadyTask(ctx, "verify-missing-session")
	item, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion one", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add checklist item: %v", err)
	}

	workDir := t.TempDir()
	reportDir := t.TempDir()

	// Successful execute without session_id — Cursor resume on should hard-fail
	// before verify invokes the runner.
	r := runnerfake.New().WithName("cursor-cli").WithoutAutoSessionID()
	hook := &hookRunner{Runner: r, preRun: func(req runner.Request) {
		cycles, _ := h.Store.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		if req.Phase == cyclesdomain.PhaseExecute {
			writeCriteriaReport(t, reportDir, cycles[0].ID, []string{item.ID})
		}
	}}
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "exec ok", []byte(`{}`), ""))

	done := h.StartHarnessRun(ctx, tsk, hook, harness.Options{
		WorkingDir: workDir,
		ReportDir:  reportDir,
	})
	h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	<-done
	cancel()

	for _, c := range r.Calls() {
		if c.Phase == cyclesdomain.PhaseVerify {
			t.Fatalf("verify must not Run when execute omitted session_id; calls=%d", len(r.Calls()))
		}
	}
}

func TestWorker_VerifyPhase_errResumeSessionNoSecondRun(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	enabled := true
	if _, err := h.Store.UpdateSettings(ctx, settingscontract.SettingsPatch{
		CursorSessionResumeEnabled: &enabled,
	}); err != nil {
		t.Fatalf("enable cursor resume: %v", err)
	}

	tsk := h.CreateReadyTask(ctx, "verify-resume-fail")
	item, err := h.Store.AddChecklistItem(ctx, tsk.ID, "criterion one", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("add checklist item: %v", err)
	}

	workDir := t.TempDir()
	reportDir := t.TempDir()
	const execSession = "sess-resume-fail"

	execDetails, err := json.Marshal(map[string]string{"session_id": execSession})
	if err != nil {
		t.Fatal(err)
	}

	r := runnerfake.New()
	hook := &hookRunner{Runner: r, preRun: func(req runner.Request) {
		cycles, _ := h.Store.ListCyclesForTask(context.Background(), req.TaskID, 1)
		if len(cycles) == 0 {
			return
		}
		if req.Phase == cyclesdomain.PhaseExecute {
			writeCriteriaReport(t, reportDir, cycles[0].ID, []string{item.ID})
		}
	}}
	r.Script(tsk.ID, cyclesdomain.PhaseExecute, runner.NewResult(
		cyclesdomain.PhaseStatusSucceeded, "exec ok", execDetails, ""))
	r.Fail(tsk.ID, cyclesdomain.PhaseVerify, runner.ErrResumeSession)

	done := h.StartHarnessRun(ctx, tsk, hook, harness.Options{
		WorkingDir: workDir,
		ReportDir:  reportDir,
	})
	h.WaitTaskStatus(ctx, tsk.ID, taskcoredomain.StatusFailed)
	<-done
	cancel()

	verifyCalls := 0
	for _, c := range r.Calls() {
		if c.Phase == cyclesdomain.PhaseVerify {
			verifyCalls++
		}
	}
	if verifyCalls != 1 {
		t.Fatalf("verify Runs = %d, want exactly 1 (no soft-fresh retry)", verifyCalls)
	}
}
