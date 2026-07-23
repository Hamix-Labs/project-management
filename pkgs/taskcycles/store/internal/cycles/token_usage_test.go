package cycles_test

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/internal/cycles"
)

func TestListPhaseTokenUsageForTask_sumsByKindAndCycle(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	tasks := taskcorestore.NewStore(db)
	cyclesStore := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, err := tasks.Create(ctx, taskcorestore.CreateTaskInput{
		Title:    "token-usage",
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

	rows, err := cycles.ListPhaseTokenUsageForTask(ctx, db, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}

	execute, verify := cyclesdomain.SumPhaseUsageByKind(rows)
	if execute.InputTokens != 100 || execute.OutputTokens != 50 {
		t.Fatalf("execute = %+v", execute)
	}
	if verify.InputTokens != 20 || verify.OutputTokens != 10 {
		t.Fatalf("verify = %+v", verify)
	}

	byCycle := cyclesdomain.SumPhaseUsageByCycleID(rows)
	got := byCycle[cycle.ID]
	if got.InputTokens != 120 || got.OutputTokens != 60 {
		t.Fatalf("cycle total = %+v", got)
	}
}
