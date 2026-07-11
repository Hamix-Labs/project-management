package store_test

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestValidateCanMarkDone_acceptsLegacyCompletions(t *testing.T) {
	t.Parallel()
	st := store.NewStore(tasktestdb.OpenSQLite(t))
	ctx := context.Background()

	tsk, err := st.Create(ctx, store.CreateTaskInput{
		Title: "t", InitialPrompt: "p", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	it, err := st.AddChecklistItem(ctx, tsk.ID, "criterion", nil, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SetChecklistItemDone(ctx, tsk.ID, it.ID, true, taskcoredomain.ActorAgent); err != nil {
		t.Fatal(err)
	}
	done := taskcoredomain.StatusDone
	if _, err := st.Update(ctx, tsk.ID, store.UpdateTaskInput{Status: &done}, taskcoredomain.ActorAgent); err != nil {
		t.Fatalf("legacy completion should allow done: %v", err)
	}
}
