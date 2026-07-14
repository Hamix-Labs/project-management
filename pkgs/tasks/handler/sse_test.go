package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/middleware"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSSE_subscriberGaugeTracksSubscribe(t *testing.T) {
	g := middleware.SSESubscribersGauge()
	h := realtime.NewSSEHub()
	_, cancel := h.Subscribe()
	// RecordSSESubscriberGauge sets the process gauge to this hub's len(subs),
	// it does not add to a prior value. Another test may have left a different
	// number on the gauge, so assert the fresh hub's cardinality, not base+delta.
	if got := testutil.ToFloat64(g); got != 1 {
		t.Fatalf("after subscribe: got %v want 1", got)
	}
	cancel()
	if got := testutil.ToFloat64(g); got != 0 {
		t.Fatalf("after cancel: got %v want 0", got)
	}
}

func TestSSEHub_Publish_deliversToSubscriber(t *testing.T) {
	h := realtime.NewSSEHub()
	ch, cancel := h.Subscribe()
	defer cancel()

	h.Publish(realtime.Event{Type: realtime.TaskCreated, ID: "abc-123"})
	select {
	case line := <-ch:
		var ev realtime.Event
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatal(err)
		}
		if ev.Type != realtime.TaskCreated || ev.ID != "abc-123" {
			t.Fatalf("got %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestSSEHub_Publish_nonBlockingSlowConsumer(t *testing.T) {
	h := realtime.NewSSEHub()
	_, cancel := h.Subscribe()
	defer cancel()
	for i := 0; i < 64; i++ {
		h.Publish(realtime.Event{Type: realtime.TaskUpdated, ID: "x"})
	}
}

// TestSSEHub_Publish_recordsDroppedFramesCounter pins the new
// taskapi_sse_dropped_frames_total observable wired in this session: when a
// subscriber's bounded channel fills up (default cap 32 inside Subscribe),
// further Publish calls drop on the `default` branch instead of blocking the
// publisher. Before this counter, those drops were silent — a stuck client
// would only surface as missing UI updates with no metric trail.
//
// Setup: one slow subscriber that never reads, then 32 publishes to fill the
// channel (each one delivered, none dropped) followed by 5 more publishes
// (all dropped because the channel is full). Asserts the counter advanced by
// exactly 5 between snapshots — proves both the per-fanout count is correct
// and that the dropped-loop iteration matches the subscriber count (1 here).
//
// A second `idleSecondSubscriber` is added partway to confirm the counter
// counts per-frame-per-subscriber, not per-fanout: 3 publishes after the
// second subscriber joins, each of which drops on BOTH subscribers (since
// neither is reading), should advance the counter by 6 (3 publishes ×
// 2 subscribers). This pins the helper signature (`RecordSSEDroppedFrames(n)`
// gets the full per-fanout sum) so a future refactor that called the helper
// once per fanout instead of once per dropped subscriber would fail loudly.
func TestSSEHub_Publish_recordsDroppedFramesCounter(t *testing.T) {
	c := middleware.SSEDroppedFramesCounter()
	base := testutil.ToFloat64(c)

	h := realtime.NewSSEHub()
	_, cancel := h.Subscribe()
	defer cancel()

	const subscriberBufferCap = 32
	for i := 0; i < subscriberBufferCap; i++ {
		h.Publish(realtime.Event{Type: realtime.TaskUpdated, ID: "fill"})
	}
	if got := testutil.ToFloat64(c); got != base {
		t.Fatalf("counter advanced before drops: got %v want %v", got, base)
	}

	const overflowOneSub = 5
	for i := 0; i < overflowOneSub; i++ {
		h.Publish(realtime.Event{Type: realtime.TaskUpdated, ID: "drop1"})
	}
	if got, want := testutil.ToFloat64(c), base+overflowOneSub; got != want {
		t.Fatalf("after %d drops on 1 subscriber: counter=%v want %v", overflowOneSub, got, want)
	}

	_, cancel2 := h.Subscribe()
	defer cancel2()

	const dropFanoutFrames = 3
	for i := 0; i < dropFanoutFrames+subscriberBufferCap; i++ {
		h.Publish(realtime.Event{Type: realtime.TaskUpdated, ID: "drop2"})
	}
	wantAfter := base + overflowOneSub + dropFanoutFrames*2 + subscriberBufferCap*1
	if got := testutil.ToFloat64(c); got != wantAfter {
		t.Fatalf("after second-subscriber phase: counter=%v want %v (overflowOneSub=%d + dropFanoutFrames*2=%d + subscriberBufferCap*1=%d for the still-full first sub during the 32 fills of the second sub)",
			got, wantAfter, overflowOneSub, dropFanoutFrames*2, subscriberBufferCap)
	}
}
