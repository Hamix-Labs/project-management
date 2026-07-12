package resume

import (
	"context"
	"encoding/json"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/verify"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestReconstructCheckpoint_readsVerifyRetryCountFromPhaseDetails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := storefake.New(t).API
	tsk, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "retry budget", InitialPrompt: "work", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	running := taskcoredomain.StatusRunning
	if _, err := st.Update(ctx, tsk.ID, taskcorestore.UpdateTaskInput{Status: &running}, taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("update: %v", err)
	}
	cycle, err := st.StartCycle(ctx, cyclescontract.StartCycleInput{TaskID: tsk.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatalf("start cycle: %v", err)
	}
	exec, err := st.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatalf("start execute: %v", err)
	}
	summary := cyclesdomain.PhaseInterruptReason
	if _, err := st.CompletePhase(ctx, cyclescontract.CompletePhaseInput{
		CycleID: cycle.ID, PhaseSeq: exec.PhaseSeq,
		Status: cyclesdomain.PhaseStatusFailed, Summary: &summary, By: taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatalf("complete execute: %v", err)
	}
	verifyPhase, err := st.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseVerify, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatalf("start verify: %v", err)
	}
	details := verify.EncodePhaseDetails(1, nil, nil, verify.PhaseDetailsOpts{VerifyRetryCount: 2})
	if _, err := st.CompletePhase(ctx, cyclescontract.CompletePhaseInput{
		CycleID: cycle.ID, PhaseSeq: verifyPhase.PhaseSeq,
		Status: cyclesdomain.PhaseStatusFailed, Summary: &summary, Details: details, By: taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatalf("complete verify: %v", err)
	}

	svc := NewService(st, Options{})
	cp, err := svc.ReconstructCheckpoint(ctx, cycle)
	if err != nil {
		t.Fatalf("checkpoint: %v", err)
	}
	if cp.VerifyAttempt != 2 {
		t.Fatalf("verifyAttempt = %d, want 2 from phase details", cp.VerifyAttempt)
	}
}

func TestParseVerifyRetryCount_roundTrip(t *testing.T) {
	t.Parallel()
	raw := verify.EncodePhaseDetails(3, nil, nil, verify.PhaseDetailsOpts{VerifyRetryCount: 1})
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := wire["verify_retry_count"]; !ok {
		t.Fatal("verify_retry_count missing from wire JSON")
	}
}
