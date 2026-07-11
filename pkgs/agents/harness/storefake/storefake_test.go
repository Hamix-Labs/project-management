package storefake_test

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func TestFake_satisfiesHarnessStore_createAndGet(t *testing.T) {
	t.Parallel()
	f := storefake.New(t)
	ctx := context.Background()
	tsk, err := f.Create(ctx, store.CreateTaskInput{
		Title:         "x",
		InitialPrompt: "p",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	got, err := f.Get(ctx, tsk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "x" {
		t.Fatalf("title = %q", got.Title)
	}
}
