package store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

func TestStore_ApplyDevTaskRowMirror_status(t *testing.T) {
	s := NewStore(tasktestdb.OpenSQLite(t))
	ctx := context.Background()
	tsk, err := s.Create(ctx, CreateTaskInput{Priority: taskcoredomain.PriorityMedium, Title: "t"}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]string{"from": "ready", "to": "running"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ApplyDevTaskRowMirror(ctx, tsk.ID, taskeventsdomain.EventStatusChanged, data); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx, tsk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != taskcoredomain.StatusRunning {
		t.Fatalf("status %q", got.Status)
	}
}

func TestStore_ListDevsimTasks_like(t *testing.T) {
	s := NewStore(tasktestdb.OpenSQLite(t))
	ctx := context.Background()
	if _, err := s.Create(ctx, CreateTaskInput{ID: "hamix-devsim-aa", Title: "a", Priority: taskcoredomain.PriorityMedium}, taskcoredomain.ActorUser); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create(ctx, CreateTaskInput{ID: "other-id", Title: "b", Priority: taskcoredomain.PriorityMedium}, taskcoredomain.ActorUser); err != nil {
		t.Fatal(err)
	}
	got, err := s.ListDevsimTasks(ctx, "hamix-devsim-%")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "hamix-devsim-aa" {
		t.Fatalf("got %+v", got)
	}
}
