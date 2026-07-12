package agentworker

import (
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func TestPublishEventNonBlocking_returnsQuicklyOnSlowPublisher(t *testing.T) {
	t.Parallel()
	pub := newBlockingPublisher()
	start := time.Now()
	go publishEventNonBlocking(pub, nil, "cycle_change", realtime.Event{
		Type: realtime.TaskCycleChanged,
		ID:   "t1",
	})
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("publishEventNonBlocking blocked caller for %v", elapsed)
	}
	close(pub.block)
}

func TestRunProgressSSEAdapter_returnsWithoutBlocking(t *testing.T) {
	t.Parallel()
	pub := newBlockingPublisher()
	adapter := newRunProgressSSEAdapter(pub, 0, nil)
	start := time.Now()
	adapter.PublishRunProgress("t1", "c1", 1, "run-1", runner.ProgressEvent{Kind: "tool"})
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("PublishRunProgress blocked for %v", elapsed)
	}
	close(pub.block)
}
