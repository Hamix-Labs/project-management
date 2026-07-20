package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checkliststore "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
)

func TestUpsertCriteriaAndVerifyReports_idempotentAndValidation(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	tasks := taskcorestore.NewStore(db)
	checklist := checkliststore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task := mustCreateReadyTask(t, tasks, "reports")
	c1, err := checklist.AddChecklistItem(ctx, task.ID, "criterion one", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := checklist.AddChecklistItem(ctx, task.ID, "criterion two", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := cycles.StartCycle(ctx, cyclesstore.StartCycleInput{
		TaskID:      task.ID,
		TriggeredBy: taskcoredomain.ActorAgent,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := cycles.UpsertCriteriaReports(ctx, cycle.ID, cycle.AttemptSeq, []cyclesstore.CriteriaReportEntry{
		{CriterionID: c1.ID, ClaimedDone: true, Evidence: "first"},
		{CriterionID: c2.ID, ClaimedDone: false, Evidence: "nope"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := cycles.UpsertCriteriaReports(ctx, cycle.ID, cycle.AttemptSeq, []cyclesstore.CriteriaReportEntry{
		{CriterionID: c1.ID, ClaimedDone: false, Evidence: "updated"},
	}); err != nil {
		t.Fatal(err)
	}

	got, err := cycles.GetCriteriaReport(ctx, cycle.ID, cycle.AttemptSeq, c1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ClaimedDone || got.Evidence != "updated" {
		t.Fatalf("criteria upsert = %+v", got)
	}
	listed, err := cycles.ListCriteriaReportsForCycle(ctx, cycle.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 {
		t.Fatalf("list criteria len = %d, want 2", len(listed))
	}

	err = cycles.UpsertCriteriaReports(ctx, cycle.ID, cycle.AttemptSeq, []cyclesstore.CriteriaReportEntry{
		{CriterionID: c1.ID, ClaimedDone: true},
		{CriterionID: c1.ID, ClaimedDone: false},
	})
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("duplicate criterion: err = %v, want ErrInvalidInput", err)
	}
	err = cycles.UpsertCriteriaReports(ctx, "", 1, []cyclesstore.CriteriaReportEntry{{CriterionID: c1.ID}})
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("empty cycle_id: err = %v, want ErrInvalidInput", err)
	}
	err = cycles.UpsertCriteriaReports(ctx, cycle.ID, 0, []cyclesstore.CriteriaReportEntry{{CriterionID: c1.ID}})
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("bad attempt_seq: err = %v, want ErrInvalidInput", err)
	}

	if err := cycles.UpsertVerifyReports(ctx, cycle.ID, cycle.AttemptSeq, []cyclesstore.VerifyReportEntry{
		{
			CriterionID:  c1.ID,
			Verified:     true,
			VerifierKind: checklistdomain.VerifierDeterministicCheck,
			Reasoning:    "ok",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := cycles.UpsertVerifyReports(ctx, cycle.ID, cycle.AttemptSeq, []cyclesstore.VerifyReportEntry{
		{
			CriterionID:  c1.ID,
			Verified:     false,
			VerifierKind: checklistdomain.VerifierAgentSelf,
			Reasoning:    "revised",
		},
	}); err != nil {
		t.Fatal(err)
	}
	verifyRows, err := cycles.ListVerifyReportsForCycle(ctx, cycle.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(verifyRows) != 1 || verifyRows[0].Verified || verifyRows[0].Reasoning != "revised" {
		t.Fatalf("verify upsert = %+v", verifyRows)
	}
}

func TestAppendCycleStreamEvent_seqAndPagination(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	tasks := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task := mustCreateReadyTask(t, tasks, "stream")
	cycle, err := cycles.StartCycle(ctx, cyclesstore.StartCycleInput{
		TaskID:      task.ID,
		TriggeredBy: taskcoredomain.ActorAgent,
	})
	if err != nil {
		t.Fatal(err)
	}
	phase, err := cycles.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}

	first, err := cycles.AppendCycleStreamEvent(ctx, cyclesstore.AppendCycleStreamEventInput{
		TaskID:   task.ID,
		CycleID:  cycle.ID,
		PhaseSeq: phase.PhaseSeq,
		Source:   "runner",
		Kind:     "progress",
		Message:  "one",
		Payload:  []byte(`{}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := cycles.AppendCycleStreamEvent(ctx, cyclesstore.AppendCycleStreamEventInput{
		TaskID:   task.ID,
		CycleID:  cycle.ID,
		PhaseSeq: phase.PhaseSeq,
		Source:   "runner",
		Kind:     "progress",
		Message:  "two",
		Payload:  []byte(`{"n":2}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.StreamSeq != 1 || second.StreamSeq != 2 {
		t.Fatalf("stream seq = %d,%d want 1,2", first.StreamSeq, second.StreamSeq)
	}

	all, err := cycles.ListCycleStreamEvents(ctx, cycle.ID, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("list all len = %d, want 2", len(all))
	}
	page, err := cycles.ListCycleStreamEvents(ctx, cycle.ID, first.StreamSeq, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(page) != 1 || page[0].StreamSeq != second.StreamSeq || page[0].Message != "two" {
		t.Fatalf("after_seq page = %+v", page)
	}

	_, err = cycles.AppendCycleStreamEvent(ctx, cyclesstore.AppendCycleStreamEventInput{
		TaskID:   "other-task",
		CycleID:  cycle.ID,
		PhaseSeq: phase.PhaseSeq,
		Source:   "runner",
		Kind:     "progress",
		Payload:  []byte(`{}`),
	})
	if !errors.Is(err, taskcoredomain.ErrNotFound) {
		t.Fatalf("wrong task binding: err = %v, want ErrNotFound", err)
	}

	_, err = cycles.AppendCycleStreamEvent(ctx, cyclesstore.AppendCycleStreamEventInput{
		TaskID:  task.ID,
		CycleID: cycle.ID,
		Source:  "runner",
		Kind:    "progress",
		Payload: []byte(`{}`),
	})
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("missing phase_seq: err = %v, want ErrInvalidInput", err)
	}
}
