package realtime

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/middleware"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSSEHub_Publish_assignsMonotonicIDs(t *testing.T) {
	h := NewSSEHubWith(SSEHubOptions{RingSize: 16, SubscriberBuffer: 16})
	const n = 50
	for i := 0; i < n; i++ {
		h.Publish(Event{Type: TaskUpdated, ID: fmt.Sprintf("t-%d", i)})
	}
	if got, want := h.LastEventID(), uint64(n); got != want {
		t.Fatalf("LastEventID=%d want %d", got, want)
	}
}

// TestSSEHub_Publish_replayFromLastEventID verifies the ring-buffer
// replay path used by reconnecting EventSource clients. After 5
// publishes a fresh subscriber requesting Last-Event-ID=2 receives
// exactly events 3, 4, 5 in publish order — no gap directive, no
// stale frames mixed in. This is the foundation of "lossless SSE":
// a client whose connection blipped at id=2 reconnects, hands its
// last seen id back, and the hub replays the missing tail.
func TestSSEHub_Publish_replayFromLastEventID(t *testing.T) {
	h := NewSSEHubWith(SSEHubOptions{RingSize: 16, SubscriberBuffer: 16})
	for i := 1; i <= 5; i++ {
		h.Publish(Event{Type: TaskUpdated, ID: fmt.Sprintf("t-%d", i)})
	}

	sub, replay, hadGap, cancel := h.SubscribeSince(2)
	defer cancel()

	if hadGap {
		t.Fatalf("expected no gap when sinceID=2 is inside the 5-event ring")
	}
	if got := len(replay); got != 3 {
		t.Fatalf("replay length=%d want 3 (events 3,4,5)", got)
	}
	if replay[0].ID != 3 || replay[1].ID != 4 || replay[2].ID != 5 {
		t.Fatalf("replay ids=[%d %d %d] want [3 4 5]", replay[0].ID, replay[1].ID, replay[2].ID)
	}

	// The new subscriber's live channel must NOT also receive the
	// replayed events — they're delivered exactly once, via the
	// snapshot return value, so the writer can flush them in order
	// before entering the heartbeat/live select.
	select {
	case ev := <-sub.Events:
		t.Fatalf("subscriber got unexpected live event during replay: id=%d line=%s", ev.ID, ev.Line)
	case <-time.After(50 * time.Millisecond):
	}
}

// TestSSEHub_Publish_gapDetectionForOldLastEventID verifies that a
// reconnecting client whose Last-Event-ID is older than the oldest
// retained ring entry triggers `hadGap=true`, which the HTTP handler
// translates into a single `{"type":"resync"}` directive. Without
// this branch the client would silently miss every event between
// "their last id" and "oldest still in the ring" — exactly the
// loss-mode the Phase 2 work is closing.
func TestSSEHub_Publish_gapDetectionForOldLastEventID(t *testing.T) {
	h := NewSSEHubWith(SSEHubOptions{RingSize: 4, SubscriberBuffer: 16})
	// 6 publishes into a 4-entry ring → oldest retained id is 3,
	// ids 1 and 2 are evicted.
	for i := 1; i <= 6; i++ {
		h.Publish(Event{Type: TaskUpdated, ID: fmt.Sprintf("t-%d", i)})
	}

	_, _, hadGap, cancel := h.SubscribeSince(1) // client says it last saw id=1
	defer cancel()
	if !hadGap {
		t.Fatalf("expected hadGap=true (sinceID=1 is older than oldest retained id=3)")
	}

	_, _, hadGapInside, cancel2 := h.SubscribeSince(3) // client says it last saw id=3 (still in ring)
	defer cancel2()
	if hadGapInside {
		t.Fatalf("expected hadGap=false (sinceID=3 is the oldest retained id)")
	}
}

// TestSSEHub_Publish_coalescesIdenticalFrames pins the 50ms dedup
// window for `{type,id}`-identical frames. A burst of 10 identical
// `task_updated:foo` publishes inside 50ms collapses to ONE wire
// frame; the coalesced counter records the other 9 drops. Cycle
// frames carry a distinct cycle_id and are intentionally NOT
// coalesced — they're tested in the next subtest.
func TestSSEHub_Publish_coalescesIdenticalFrames(t *testing.T) {
	c := middleware.SSECoalescedCounter()
	base := testutil.ToFloat64(c)

	h := NewSSEHubWith(SSEHubOptions{
		RingSize:         16,
		SubscriberBuffer: 32,
		CoalesceWindow:   50 * time.Millisecond,
	})
	ch, cancel := h.Subscribe()
	defer cancel()

	for i := 0; i < 10; i++ {
		h.Publish(Event{Type: TaskUpdated, ID: "foo"})
	}

	// Drain everything that landed within 100ms — only the first
	// publish should have made it through; the other 9 collapsed
	// before fanout.
	delivered := 0
	timeout := time.After(100 * time.Millisecond)
drain:
	for {
		select {
		case <-ch:
			delivered++
		case <-timeout:
			break drain
		}
	}
	if delivered != 1 {
		t.Fatalf("delivered=%d want 1 (other 9 should coalesce)", delivered)
	}
	if got, want := testutil.ToFloat64(c), base+9; got != want {
		t.Fatalf("coalesced counter=%v want %v", got, want)
	}
}

// TestSSEHub_Publish_doesNotCoalesceCycleFrames pins the
// "cycle frames are informationally distinct" rule documented in the
// hub's coalesceKey: each cycle phase transition carries a different
// cycle_id, so the SPA needs every frame to refresh the right slot
// on the task detail page even when several land inside the 50ms
// window.
func TestSSEHub_Publish_doesNotCoalesceCycleFrames(t *testing.T) {
	h := NewSSEHubWith(SSEHubOptions{
		RingSize:         16,
		SubscriberBuffer: 32,
		CoalesceWindow:   50 * time.Millisecond,
	})
	ch, cancel := h.Subscribe()
	defer cancel()

	h.Publish(Event{Type: TaskCycleChanged, ID: "task-1", CycleID: "c-1"})
	h.Publish(Event{Type: TaskCycleChanged, ID: "task-1", CycleID: "c-1"})
	h.Publish(Event{Type: TaskCycleChanged, ID: "task-1", CycleID: "c-1"})

	delivered := 0
	timeout := time.After(100 * time.Millisecond)
drain:
	for {
		select {
		case <-ch:
			delivered++
		case <-timeout:
			break drain
		}
	}
	if delivered != 3 {
		t.Fatalf("delivered=%d want 3 (cycle frames must NOT coalesce)", delivered)
	}
}

// TestSSEHub_Publish_evictsSlowConsumer pins the
// "overflow → evict + resync" backpressure contract. A subscriber
// whose channel fills up is removed from the registration set and
// its cancel channel is closed (the writer goroutine then sends a
// resync directive on the wire and shuts the HTTP stream down).
// This is loss-free under Last-Event-ID resume: the client
// reconnects with its last-seen id and replays from the ring.
func TestSSEHub_Publish_evictsSlowConsumer(t *testing.T) {
	c := middleware.SSESubscriberEvictionsCounter()
	base := testutil.ToFloat64(c)

	h := NewSSEHubWith(SSEHubOptions{
		RingSize:         128,
		SubscriberBuffer: 4,
	})
	sub, _, _, cancel := h.SubscribeSince(0)
	defer cancel()

	// First 4 publishes fill the per-subscriber buffer (no overflow).
	for i := 0; i < 4; i++ {
		h.Publish(Event{Type: TaskUpdated, ID: fmt.Sprintf("t-%d", i)})
	}
	// Next publish overflows → eviction.
	h.Publish(Event{Type: TaskUpdated, ID: "overflow"})

	select {
	case <-sub.Cancelled:
		// Expected: the hub closed our cancel channel as part of eviction.
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected sub.Cancelled to be closed after overflow")
	}
	if got, want := testutil.ToFloat64(c), base+1; got != want {
		t.Fatalf("eviction counter=%v want %v", got, want)
	}
}

// TestHTTP_SSE_emitsIDLineForEventSourceResume verifies the on-the-wire
// frame shape every event ships with, so the browser EventSource
// captures `id: N` as Last-Event-ID for reconnect. Older deployments
// shipped only `data: ...` — a client whose connection blipped would
// reconnect with no Last-Event-ID header and silently miss every
// in-flight event. The plan calls this out as the critical wire
// contract behind lossless SSE.
func TestSSEHub_Publish_concurrentSafetyUnderLoad(t *testing.T) {
	h := NewSSEHubWith(SSEHubOptions{
		RingSize:         128,
		SubscriberBuffer: 64,
		CoalesceWindow:   1 * time.Millisecond,
	})

	var wg sync.WaitGroup
	const subs = 8
	const publishersPerSub = 4
	const eventsPerPublisher = 200

	for i := 0; i < subs; i++ {
		ch, cancel := h.Subscribe()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer cancel()
			deadline := time.After(2 * time.Second)
			for {
				select {
				case <-ch:
				case <-deadline:
					return
				}
			}
		}()
	}

	var pub sync.WaitGroup
	for p := 0; p < publishersPerSub; p++ {
		pub.Add(1)
		go func(idx int) {
			defer pub.Done()
			for i := 0; i < eventsPerPublisher; i++ {
				h.Publish(Event{
					Type: TaskUpdated,
					ID:   fmt.Sprintf("p%d-%d", idx, i),
				})
			}
		}(p)
	}
	pub.Wait()
	wg.Wait()

	if h.LastEventID() == 0 {
		t.Fatalf("LastEventID should advance under concurrent publish")
	}
}
