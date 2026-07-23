package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	checkliststore "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
)

func TestRequestTaskPolish_fromReviewSetsReadyAndPending(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, cycleID := mustReviewTaskWithSucceededCycle(t, st, cycles)

	updated, prev, err := st.RequestTaskPolish(ctx, taskcorestore.RequestPolishInput{
		TaskID:       task.ID,
		Instructions: "tighten commit messages",
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if prev != domain.StatusReview {
		t.Fatalf("prev = %q, want review", prev)
	}
	if updated.Status != domain.StatusReady {
		t.Fatalf("status = %q, want ready", updated.Status)
	}
	pr := updated.PendingRetry
	if pr == nil || pr.NormalizeKind() != domain.PendingKindPolish || pr.Mode != domain.RetryResume ||
		pr.ParentCycleID != cycleID || pr.Instructions != "tighten commit messages" {
		t.Fatalf("PendingRetry = %+v", pr)
	}

	stored, err := st.Get(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.PendingRetry == nil || !stored.PendingRetry.Equal(updated.PendingRetry) {
		t.Fatalf("stored PendingRetry = %+v, want %+v", stored.PendingRetry, updated.PendingRetry)
	}
}

func TestRequestTaskPolish_idempotentSameIntent(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, _ := mustReviewTaskWithSucceededCycle(t, st, cycles)
	in := taskcorestore.RequestPolishInput{
		TaskID:       task.ID,
		Instructions: "polish UI spacing",
	}
	first, _, err := st.RequestTaskPolish(ctx, in, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	second, prev, err := st.RequestTaskPolish(ctx, in, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if prev != domain.StatusReady {
		t.Fatalf("prev = %q, want ready", prev)
	}
	if second.PendingRetry == nil || !second.PendingRetry.Equal(first.PendingRetry) {
		t.Fatalf("second PendingRetry = %+v, want %+v", second.PendingRetry, first.PendingRetry)
	}
}

func TestRequestTaskPolish_conflictDifferentIntent(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, _ := mustReviewTaskWithSucceededCycle(t, st, cycles)
	if _, _, err := st.RequestTaskPolish(ctx, taskcorestore.RequestPolishInput{
		TaskID: task.ID, Instructions: "first",
	}, domain.ActorUser); err != nil {
		t.Fatal(err)
	}
	_, _, err := st.RequestTaskPolish(ctx, taskcorestore.RequestPolishInput{
		TaskID: task.ID, Instructions: "second",
	}, domain.ActorUser)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestRequestTaskPolish_rejectsNonReview(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "ready", Status: domain.StatusReady, Priority: domain.PriorityMedium,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = st.RequestTaskPolish(ctx, taskcorestore.RequestPolishInput{
		TaskID: task.ID, Instructions: "nope",
	}, domain.ActorUser)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestRequestTaskPolish_rejectsEmptyInstructions(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, _ := mustReviewTaskWithSucceededCycle(t, st, cycles)
	_, _, err := st.RequestTaskPolish(ctx, taskcorestore.RequestPolishInput{
		TaskID: task.ID, Instructions: "   ",
	}, domain.ActorUser)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestRequestTaskPolish_rejectsAgent(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, _ := mustReviewTaskWithSucceededCycle(t, st, cycles)
	_, _, err := st.RequestTaskPolish(ctx, taskcorestore.RequestPolishInput{
		TaskID: task.ID, Instructions: "polish",
	}, domain.ActorAgent)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestRequestTaskPolish_rejectsFailedParentCycle(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	ctx := context.Background()

	task, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		Title: "review", Status: domain.StatusReview, Priority: domain.PriorityMedium,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := cycles.StartCycle(ctx, cyclesstore.StartCycleInput{
		TaskID: task.ID, TriggeredBy: domain.ActorAgent,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cycles.TerminateCycle(ctx, cycle.ID, cyclesdomain.CycleStatusFailed, "boom", domain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	_, _, err = st.RequestTaskPolish(ctx, taskcorestore.RequestPolishInput{
		TaskID: task.ID, Instructions: "polish", ParentCycleID: cycle.ID,
	}, domain.ActorUser)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestRequestTaskPolish_flagsAndNewCriteria(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	st := taskcorestore.NewStore(db)
	cycles := cyclesstore.NewStore(db)
	checklist := checkliststore.NewStore(db)
	ctx := context.Background()

	task, _ := mustReviewTaskWithSucceededCycle(t, st, cycles)
	itemA, err := checklist.AddChecklistItem(ctx, task.ID, "Auth works", nil, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	itemB, err := checklist.AddChecklistItem(ctx, task.ID, "Tests pass", nil, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := checklist.SetChecklistItemDone(ctx, task.ID, itemA.ID, true, domain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	if err := checklist.SetChecklistItemDone(ctx, task.ID, itemB.ID, true, domain.ActorAgent); err != nil {
		t.Fatal(err)
	}

	updated, _, err := st.RequestTaskPolish(ctx, taskcorestore.RequestPolishInput{
		TaskID:              task.ID,
		Instructions:        "fix auth",
		FlaggedCriterionIDs: []string{itemA.ID},
		NewCriteria:         []string{"Docs updated"},
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	pr := updated.PendingRetry
	if pr == nil || pr.SkipVerify || len(pr.FlaggedCriterionIDs) != 1 || pr.FlaggedCriterionIDs[0] != itemA.ID {
		t.Fatalf("PendingRetry = %+v", pr)
	}
	if len(pr.NewCriterionIDs) != 1 {
		t.Fatalf("NewCriterionIDs = %#v", pr.NewCriterionIDs)
	}

	items, err := checklist.ListChecklistForSubject(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("items = %d, want 3", len(items))
	}
	var aDone, bDone, newFound bool
	for _, it := range items {
		switch it.ID {
		case itemA.ID:
			aDone = it.Done
		case itemB.ID:
			bDone = it.Done
		case pr.NewCriterionIDs[0]:
			newFound = true
			if it.Done {
				t.Fatal("new criterion should not be done")
			}
		}
	}
	if aDone {
		t.Fatal("flagged criterion should be reopened")
	}
	if !bDone {
		t.Fatal("unflagged criterion should stay done")
	}
	if !newFound {
		t.Fatal("new criterion missing")
	}
}
