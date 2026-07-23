package harness

import (
	"context"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestPersistProgress_succeedsWithExpiredCallerContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := storefake.New(t)

	tsk, err := st.API.Create(ctx, taskcorestore.CreateTaskInput{
		Title:         "progress-expired-ctx",
		InitialPrompt: "work",
		Priority:      taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	cycle, err := st.API.StartCycle(ctx, cyclescontract.StartCycleInput{
		TaskID:      tsk.ID,
		TriggeredBy: taskcoredomain.ActorAgent,
	})
	if err != nil {
		t.Fatal(err)
	}
	phase, err := st.API.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		t.Fatal(err)
	}

	h := New(st, runnerfake.New(), Options{})
	expired, cancel := context.WithTimeout(ctx, time.Nanosecond)
	cancel()
	<-expired.Done()

	h.persistProgress(expired, tsk.ID, cycle.ID, phase.PhaseSeq, runner.ProgressEvent{
		Kind:    "tool_call",
		Subtype: "running",
		Tool:    "ReadFile",
		Message: "still running after caller deadline",
	})

	stream, err := st.API.ListCycleStreamEvents(ctx, cycle.ID, 0, 10)
	if err != nil {
		t.Fatalf("list stream: %v", err)
	}
	if len(stream) != 1 {
		t.Fatalf("persisted stream events: got %d want 1 (%+v)", len(stream), stream)
	}
	if stream[0].Kind != "tool_call" || stream[0].Message != "still running after caller deadline" {
		t.Fatalf("stream event = %+v", stream[0])
	}
}
