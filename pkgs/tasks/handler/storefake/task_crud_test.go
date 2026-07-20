package storefake_test

import (
	"context"
	"errors"
	"testing"

	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/storefake"
)

func TestTaskCRUDFake_implements_focused_and_composed_contracts(t *testing.T) {
	t.Parallel()
	var _ taskcorecontract.TaskGetter = (*storefake.TaskCRUDFake)(nil)
	var _ taskcorecontract.TaskReader = (*storefake.TaskCRUDFake)(nil)
	var _ taskcorecontract.TaskWriter = (*storefake.TaskCRUDFake)(nil)
	var _ taskcorecontract.TaskDepsStore = (*storefake.TaskCRUDFake)(nil)
	var _ taskcorecontract.TaskOpsStore = (*storefake.TaskCRUDFake)(nil)
	var _ taskcorecontract.TaskCRUDStore = (*storefake.TaskCRUDFake)(nil)
}

func TestHandlerStoreFake_implements_HandlerStore(t *testing.T) {
	t.Parallel()
	var _ handler.HandlerStore = (*storefake.HandlerStoreFake)(nil)
}

func TestTaskCRUDFake_RetryAndGate_recording(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := storefake.NewTaskCRUD()

	wantRetry := &taskcoredomain.Task{ID: "t1", Status: taskcoredomain.StatusReady}
	fake.OnRetry(wantRetry)
	got, err := fake.RequestTaskRetry(ctx, taskcorecontract.RequestRetryInput{
		TaskID: "t1", Mode: taskcoredomain.RetryFresh, ParentCycleID: "c1",
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("Retry: %v", err)
	}
	if got.ID != wantRetry.ID {
		t.Fatalf("retry task = %+v", got)
	}
	calls := fake.RetryCalls()
	if len(calls) != 1 || calls[0].Input.ParentCycleID != "c1" || calls[0].By != taskcoredomain.ActorUser {
		t.Fatalf("RetryCalls = %+v", calls)
	}

	fake.FailRetry(taskcoredomain.ErrConflict)
	if _, err := fake.RequestTaskRetry(ctx, taskcorecontract.RequestRetryInput{TaskID: "t1"}, taskcoredomain.ActorUser); !errors.Is(err, taskcoredomain.ErrConflict) {
		t.Fatalf("FailRetry: got %v", err)
	}

	wantGate := &taskcoredomain.Task{ID: "t1", Status: taskcoredomain.StatusReady}
	fake.OnGate(wantGate)
	got, err = fake.ApplyTaskGateAction(ctx, "t1", "release", taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("Gate: %v", err)
	}
	if got.ID != wantGate.ID {
		t.Fatalf("gate task = %+v", got)
	}
	gateCalls := fake.GateCalls()
	if len(gateCalls) != 1 || gateCalls[0].Action != "release" {
		t.Fatalf("GateCalls = %+v", gateCalls)
	}

	fake.FailGate(taskcoredomain.ErrNotFound)
	if _, err := fake.ApplyTaskGateAction(ctx, "t1", "hold", taskcoredomain.ActorUser); !errors.Is(err, taskcoredomain.ErrNotFound) {
		t.Fatalf("FailGate: got %v", err)
	}
}

func TestHandlerStoreFake_ListCyclesAndEvents_recording(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := storefake.NewHandlerStore()

	cycles := []cyclesdomain.TaskCycle{{ID: "c1", TaskID: "t1", AttemptSeq: 1}}
	fake.OnListCycles(cycles)
	gotCycles, err := fake.ListCyclesForTaskBefore(ctx, "t1", 0, 10)
	if err != nil {
		t.Fatalf("ListCycles: %v", err)
	}
	if len(gotCycles) != 1 || gotCycles[0].ID != "c1" {
		t.Fatalf("cycles = %+v", gotCycles)
	}
	cycleCalls := fake.ListCyclesCalls()
	if len(cycleCalls) != 1 || cycleCalls[0].Limit != 10 {
		t.Fatalf("ListCyclesCalls = %+v", cycleCalls)
	}

	fake.FailListCycles(taskcoredomain.ErrNotFound)
	if _, err := fake.ListCyclesForTaskBefore(ctx, "t1", 0, 5); !errors.Is(err, taskcoredomain.ErrNotFound) {
		t.Fatalf("FailListCycles: got %v", err)
	}

	events := []taskeventsdomain.TaskEvent{{TaskID: "t1", Seq: 1, Type: taskeventsdomain.EventCycleStarted}}
	fake.OnListEvents(events)
	gotEvents, err := fake.ListTaskEvents(ctx, "t1")
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(gotEvents) != 1 || gotEvents[0].Type != taskeventsdomain.EventCycleStarted {
		t.Fatalf("events = %+v", gotEvents)
	}
	eventCalls := fake.ListEventsCalls()
	if len(eventCalls) != 1 || eventCalls[0].TaskID != "t1" {
		t.Fatalf("ListEventsCalls = %+v", eventCalls)
	}

	fake.FailListEvents(taskcoredomain.ErrConflict)
	if _, err := fake.ListTaskEvents(ctx, "t1"); !errors.Is(err, taskcoredomain.ErrConflict) {
		t.Fatalf("FailListEvents: got %v", err)
	}
}
