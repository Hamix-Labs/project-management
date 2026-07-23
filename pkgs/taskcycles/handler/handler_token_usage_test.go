package handler

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
)

func TestProjectTokenUsage_splitsExecuteAndVerify(t *testing.T) {
	t.Parallel()

	exec := cyclesdomain.TokenUsage{InputTokens: 100, OutputTokens: 50}
	verify := cyclesdomain.TokenUsage{InputTokens: 20, OutputTokens: 10}
	got := projectTokenUsage(exec, verify, true)

	if got.ConsumedTokens != 180 {
		t.Fatalf("consumed_tokens = %d want 180", got.ConsumedTokens)
	}
	if got.ExecuteConsumedTokens != 150 {
		t.Fatalf("execute_consumed_tokens = %d want 150", got.ExecuteConsumedTokens)
	}
	if got.VerifyConsumedTokens != 30 {
		t.Fatalf("verify_consumed_tokens = %d want 30", got.VerifyConsumedTokens)
	}
	if got.InputTokens != 120 || got.OutputTokens != 60 {
		t.Fatalf("component totals = %+v", got)
	}
	if !got.Known {
		t.Fatal("known = false want true")
	}
}

func TestProjectTokenUsageFromRows_unknownWhenEmpty(t *testing.T) {
	t.Parallel()

	got := projectTokenUsageFromRows(nil)
	if got.Known {
		t.Fatal("known = true want false for empty rows")
	}
}

func TestProjectCycleTokenUsage_filtersByCycle(t *testing.T) {
	t.Parallel()

	cycleA := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	cycleB := "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	rows := []cyclesdomain.PhaseUsageRow{
		{CycleID: cycleA, AttemptSeq: 1, Phase: cyclesdomain.PhaseExecute, Usage: cyclesdomain.TokenUsage{TotalTokens: 100}},
		{CycleID: cycleA, AttemptSeq: 1, Phase: cyclesdomain.PhaseVerify, Usage: cyclesdomain.TokenUsage{TotalTokens: 20}},
		{CycleID: cycleB, AttemptSeq: 2, Phase: cyclesdomain.PhaseExecute, Usage: cyclesdomain.TokenUsage{TotalTokens: 50}},
	}

	gotA := projectCycleTokenUsage(rows, cycleA)
	if gotA.ConsumedTokens != 120 || gotA.ExecuteConsumedTokens != 100 || gotA.VerifyConsumedTokens != 20 {
		t.Fatalf("cycle A = %+v", gotA)
	}

	gotB := projectCycleTokenUsage(rows, cycleB)
	if gotB.ConsumedTokens != 50 || gotB.ExecuteConsumedTokens != 50 || gotB.VerifyConsumedTokens != 0 {
		t.Fatalf("cycle B = %+v", gotB)
	}
}

func TestShareOfTaskPct_nullWhenUnknownOrZero(t *testing.T) {
	t.Parallel()

	if shareOfTaskPct(10, 100, false) != nil {
		t.Fatal("expected nil when unknown")
	}
	if shareOfTaskPct(10, 0, true) != nil {
		t.Fatal("expected nil when task consumed is zero")
	}
	if shareOfTaskPct(0, 100, true) != nil {
		t.Fatal("expected nil when cycle consumed is zero")
	}

	got := shareOfTaskPct(25, 100, true)
	if got == nil || math.Abs(*got-25) > 0.0001 {
		t.Fatalf("share = %v want 25", got)
	}
}

func TestTaskCycleDetailFromDomain_attachesTokenUsageFromPhases(t *testing.T) {
	t.Parallel()

	cycle := &cyclesdomain.TaskCycle{
		ID:         "11111111-1111-4111-8111-111111111111",
		TaskID:     "22222222-2222-4222-8222-222222222222",
		AttemptSeq: 1,
		Status:     cyclesdomain.CycleStatusSucceeded,
	}
	phases := []cyclesdomain.TaskCyclePhase{
		{
			Phase:       cyclesdomain.PhaseExecute,
			DetailsJSON: json.RawMessage(`{"usage":{"inputTokens":100,"outputTokens":50}}`),
		},
		{
			Phase:       cyclesdomain.PhaseVerify,
			DetailsJSON: json.RawMessage(`{"verification":{"attempt_seq":1},"usage":{"inputTokens":20,"outputTokens":10}}`),
		},
	}

	resp := taskCycleDetailFromDomain(cycle, phases)
	if resp.TokenUsage == nil {
		t.Fatal("token_usage is nil")
	}
	if resp.TokenUsage.ConsumedTokens != 180 {
		t.Fatalf("consumed_tokens = %d want 180", resp.TokenUsage.ConsumedTokens)
	}
	if resp.TokenUsage.ExecuteConsumedTokens != 150 || resp.TokenUsage.VerifyConsumedTokens != 30 {
		t.Fatalf("phase split = %+v", resp.TokenUsage)
	}
}

func TestGetTaskTokenUsage_projectsTotalsAndShare(t *testing.T) {
	t.Parallel()

	db := tasktestdb.OpenSQLite(t)
	tasks := taskcorestore.NewStore(db)
	cyclesStore := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, err := tasks.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "token-usage-api",
		Status:   taskcoredomain.StatusReady,
		Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	cycle, err := cyclesStore.StartCycle(ctx, cyclesstore.StartCycleInput{
		TaskID:      task.ID,
		TriggeredBy: taskcoredomain.ActorAgent,
	})
	if err != nil {
		t.Fatal(err)
	}

	execPhase, err := cyclesStore.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}
	execSummary := "execute ok"
	if _, err := cyclesStore.CompletePhase(ctx, cyclesstore.CompletePhaseInput{
		CycleID:  cycle.ID,
		PhaseSeq: execPhase.PhaseSeq,
		Status:   cyclesdomain.PhaseStatusSucceeded,
		Summary:  &execSummary,
		Details:  []byte(`{"usage":{"inputTokens":100,"outputTokens":50}}`),
		By:       taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatal(err)
	}

	verifyPhase, err := cyclesStore.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseVerify, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}
	verifySummary := "verify ok"
	if _, err := cyclesStore.CompletePhase(ctx, cyclesstore.CompletePhaseInput{
		CycleID:  cycle.ID,
		PhaseSeq: verifyPhase.PhaseSeq,
		Status:   cyclesdomain.PhaseStatusSucceeded,
		Summary:  &verifySummary,
		Details:  []byte(`{"verification":{"attempt_seq":1},"usage":{"inputTokens":20,"outputTokens":10}}`),
		By:       taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	Register(mux, Deps{
		Cycles: cyclesStore,
		Tasks:  tasks,
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks/"+task.ID+"/token-usage", nil)
	req.SetPathValue("id", task.ID)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	var got taskTokenUsageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if got.TaskID != task.ID {
		t.Fatalf("task_id = %q want %q", got.TaskID, task.ID)
	}
	if !got.TokenUsage.Known || got.TokenUsage.ConsumedTokens != 180 {
		t.Fatalf("task token_usage = %+v", got.TokenUsage)
	}
	if len(got.Attempts) != 1 {
		t.Fatalf("attempts = %d want 1", len(got.Attempts))
	}
	attempt := got.Attempts[0]
	if attempt.CycleID != cycle.ID || attempt.AttemptSeq != 1 {
		t.Fatalf("attempt identity = %+v", attempt)
	}
	if attempt.TokenUsage.ConsumedTokens != 180 {
		t.Fatalf("attempt token_usage = %+v", attempt.TokenUsage)
	}
	if attempt.ShareOfTaskPct == nil || math.Abs(*attempt.ShareOfTaskPct-100) > 0.0001 {
		t.Fatalf("share_of_task_pct = %v want 100", attempt.ShareOfTaskPct)
	}
}

func TestGetTaskCycles_attachesTokenUsagePerCycle(t *testing.T) {
	t.Parallel()

	db := tasktestdb.OpenSQLite(t)
	tasks := taskcorestore.NewStore(db)
	cyclesStore := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, err := tasks.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "token-usage-list",
		Status:   taskcoredomain.StatusReady,
		Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}

	cycle, err := cyclesStore.StartCycle(ctx, cyclesstore.StartCycleInput{
		TaskID:      task.ID,
		TriggeredBy: taskcoredomain.ActorAgent,
	})
	if err != nil {
		t.Fatal(err)
	}

	execPhase, err := cyclesStore.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}
	execSummary := "execute ok"
	if _, err := cyclesStore.CompletePhase(ctx, cyclesstore.CompletePhaseInput{
		CycleID:  cycle.ID,
		PhaseSeq: execPhase.PhaseSeq,
		Status:   cyclesdomain.PhaseStatusSucceeded,
		Summary:  &execSummary,
		Details:  []byte(`{"usage":{"totalTokens":150}}`),
		By:       taskcoredomain.ActorAgent,
	}); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	Register(mux, Deps{
		Cycles: cyclesStore,
		Tasks:  tasks,
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks/"+task.ID+"/cycles", nil)
	req.SetPathValue("id", task.ID)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	var got taskCyclesListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if len(got.Cycles) != 1 {
		t.Fatalf("cycles = %d want 1", len(got.Cycles))
	}
	if got.Cycles[0].TokenUsage == nil {
		t.Fatal("token_usage is nil on listed cycle")
	}
	if got.Cycles[0].TokenUsage.ConsumedTokens != 150 || got.Cycles[0].TokenUsage.ExecuteConsumedTokens != 150 {
		t.Fatalf("token_usage = %+v", got.Cycles[0].TokenUsage)
	}
}
