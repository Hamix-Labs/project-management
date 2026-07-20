package realtime

import (
	"encoding/json"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/middleware"
)

type BufferedEvent struct {
	ID   uint64
	Line string
	At   time.Time
}

type Subscriber struct {
	Events    chan BufferedEvent
	legacy    chan string
	Cancelled chan struct{}
}

// SSEHub fans out task change notifications to all connected SSE clients.
//
// The hub keeps a bounded ring buffer of recent events keyed by a
// monotonically increasing event id. The HTTP handler honors the
// browser's Last-Event-ID header on reconnect by replaying the tail of
// the ring; if the requested id is older than the oldest retained
// event the handler emits one `{"type":"resync"}` frame so the client
// drops every cache and starts over from REST. Slow consumers are
// evicted (their cancel channel is closed) instead of silently dropping
// frames — they reconnect, replay via Last-Event-ID, and either catch
// up or fall through to the same resync directive. End result: the
// /events stream is semantically lossless even when individual TCP
// connections stutter.
//
// Coalescing: identical {Type, ID} task / settings frames published
// inside the coalesceWindow collapse to one wire frame so a burst of
// supervisor reloads or duplicate task_updated events does not spam
// the fanout. Cycle frames carry a distinct cycle_id and are not
// coalesced.
type SSEHub struct {
	mu              sync.Mutex
	subs            map[*Subscriber]struct{}
	ring            []BufferedEvent
	ringHead        int
	ringFilled      bool
	nextID          atomic.Uint64
	coalesceWindow  time.Duration
	lastEmitted     map[string]coalesceEntry
	subBuf          int
	heartbeatPeriod time.Duration
}

type coalesceEntry struct {
	at time.Time
}

// SSEHubOptions tunes the hub. Zero values pick safe production
// defaults; tests override individual fields to exercise edge cases.
type SSEHubOptions struct {
	RingSize         int
	SubscriberBuffer int
	CoalesceWindow   time.Duration
	HeartbeatPeriod  time.Duration
}

// DefaultSSEHubOptions are the production-grade defaults documented in
// the architecture plan: 1024-event ring, 256-frame per-subscriber
// buffer, 50ms coalescing window, 15s heartbeats.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func DefaultSSEHubOptions() SSEHubOptions {
	return SSEHubOptions{
		RingSize:         1024,
		SubscriberBuffer: 256,
		CoalesceWindow:   50 * time.Millisecond,
		HeartbeatPeriod:  15 * time.Second,
	}
}

// NewSSEHub returns a hub with no subscribers wired to test-friendly
// defaults: ring + per-subscriber buffer + heartbeats are kept (they
// are loss-prevention, not behavioral changes) but coalescing is OFF
// so back-to-back distinct operations never collide on the 50ms
// dedup window in unit tests. Production wiring in
// `cmd/taskapi/run_helpers.go` opts in to the production defaults via
// `NewSSEHubWith(DefaultSSEHubOptions())`. It is safe for concurrent
// use.
func NewSSEHub() *SSEHub {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "realtime.NewSSEHub")
	opts := DefaultSSEHubOptions()
	opts.CoalesceWindow = 0
	return NewSSEHubWith(opts)
}

// NewSSEHubWith builds a hub with caller-supplied tuning. Invalid values
// (zero or negative) fall back to the matching DefaultSSEHubOptions
// field so callers can override one knob at a time.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func NewSSEHubWith(opts SSEHubOptions) *SSEHub {
	d := DefaultSSEHubOptions()
	if opts.RingSize <= 0 {
		opts.RingSize = d.RingSize
	}
	if opts.SubscriberBuffer <= 0 {
		opts.SubscriberBuffer = d.SubscriberBuffer
	}
	if opts.CoalesceWindow < 0 {
		opts.CoalesceWindow = 0
	}
	if opts.HeartbeatPeriod < 0 {
		opts.HeartbeatPeriod = 0
	}
	return &SSEHub{
		subs:            make(map[*Subscriber]struct{}),
		ring:            make([]BufferedEvent, opts.RingSize),
		coalesceWindow:  opts.CoalesceWindow,
		lastEmitted:     make(map[string]coalesceEntry),
		subBuf:          opts.SubscriberBuffer,
		heartbeatPeriod: opts.HeartbeatPeriod,
	}
}

// SubscribeSince registers a live subscriber and optionally returns ring
// replay since sinceID. HadGap means the client must resync via REST.
func (h *SSEHub) SubscribeSince(sinceID uint64) (sub *Subscriber, replay []BufferedEvent, hadGap bool, cancel func()) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "realtime.SSEHub.SubscribeSince", "since_id", sinceID)
	sub = &Subscriber{
		Events:    make(chan BufferedEvent, h.subBuf),
		Cancelled: make(chan struct{}),
	}
	h.mu.Lock()
	if sinceID > 0 {
		replay, hadGap = h.snapshotSinceLocked(sinceID)
	}
	h.subs[sub] = struct{}{}
	n := len(h.subs)
	h.mu.Unlock()
	middleware.RecordSSESubscriberGauge(n)
	cancel = func() {
		h.mu.Lock()
		delete(h.subs, sub)
		n := len(h.subs)
		h.mu.Unlock()
		middleware.RecordSSESubscriberGauge(n)
	}
	return sub, replay, hadGap, cancel
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *SSEHub) snapshotSinceLocked(sinceID uint64) (out []BufferedEvent, hadGap bool) {
	if !h.ringFilled && h.ringHead == 0 {
		return nil, false
	}
	size := len(h.ring)
	start := 0
	count := h.ringHead
	if h.ringFilled {
		start = h.ringHead
		count = size
	}
	out = make([]BufferedEvent, 0, count)
	var oldestID uint64
	for i := 0; i < count; i++ {
		e := h.ring[(start+i)%size]
		if i == 0 {
			oldestID = e.ID
		}
		if e.ID > sinceID {
			out = append(out, e)
		}
	}
	if sinceID < oldestID-1 || (sinceID == 0 && oldestID > 1) {
		hadGap = true
	}
	return out, hadGap
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *SSEHub) appendRingLocked(ev BufferedEvent) {
	h.ring[h.ringHead] = ev
	h.ringHead++
	if h.ringHead == len(h.ring) {
		h.ringHead = 0
		h.ringFilled = true
	}
}

// Publish delivers one JSON-encoded event to all current subscribers.
func (h *SSEHub) Publish(ev Event) {
	if h == nil {
		return
	}
	now := time.Now()
	if h.coalesceWindow > 0 {
		key := CoalesceKey(ev)
		if key != "" {
			h.mu.Lock()
			last, seen := h.lastEmitted[key]
			if seen && now.Sub(last.at) < h.coalesceWindow {
				h.mu.Unlock()
				middleware.RecordSSECoalesced(1)
				slog.Debug("sse coalesce drop",
					"cmd", calltrace.LogCmd, "operation", "tasks.sse.publish",
					"event_type", ev.Type, "task_id", ev.ID,
					"window_ms", h.coalesceWindow.Milliseconds())
				return
			}
			h.lastEmitted[key] = coalesceEntry{at: now}
			h.mu.Unlock()
		}
	}
	b, err := json.Marshal(ev)
	if err != nil {
		slog.Error("sse publish marshal failed", "cmd", calltrace.LogCmd, "operation", "tasks.sse.publish", "err", err)
		return
	}
	middleware.RecordSSEPublish()
	id := h.nextID.Add(1)
	be := BufferedEvent{ID: id, Line: string(b), At: now}

	h.mu.Lock()
	h.appendRingLocked(be)
	out := make([]*Subscriber, 0, len(h.subs))
	for s := range h.subs {
		out = append(out, s)
	}
	var ringOldestAge time.Duration
	if len(h.ring) > 0 {
		if age := now.Sub(h.ring[0].At); age > 0 {
			ringOldestAge = age
		}
	}
	h.mu.Unlock()

	dropped := 0
	evicted := 0
	for _, s := range out {
		if s.legacy != nil {
			select {
			case s.legacy <- be.Line:
			default:
				dropped++
			}
			continue
		}
		select {
		case s.Events <- be:
			if len(s.Events) <= 1 {
				middleware.RecordSSESubscriberLag(0)
			}
		default:
			if ringOldestAge > 0 {
				middleware.RecordSSESubscriberLag(ringOldestAge.Seconds())
			}
			h.evictSubscriber(s)
			evicted++
			dropped++
		}
	}
	if dropped > 0 {
		middleware.RecordSSEDroppedFrames(dropped)
	}
	if evicted > 0 {
		middleware.RecordSSESubscriberEvictions(evicted)
		slog.Warn("sse evicted slow subscribers",
			"cmd", calltrace.LogCmd, "operation", "tasks.sse.publish",
			"event_type", ev.Type, "task_id", ev.ID,
			"event_id", id, "evicted", evicted, "remaining", len(out)-evicted)
	}
	if len(out) > 0 {
		slog.Debug("sse fanout", "cmd", calltrace.LogCmd, "operation", "tasks.sse.publish",
			"event_type", ev.Type, "task_id", ev.ID, "event_id", id,
			"subscribers", len(out), "dropped", dropped, "evicted", evicted)
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *SSEHub) evictSubscriber(s *Subscriber) {
	h.mu.Lock()
	if _, ok := h.subs[s]; ok {
		delete(h.subs, s)
		close(s.Cancelled)
	}
	n := len(h.subs)
	h.mu.Unlock()
	middleware.RecordSSESubscriberGauge(n)
}

// LastEventID returns the highest event id allocated by the hub.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *SSEHub) LastEventID() uint64 {
	if h == nil {
		return 0
	}
	return h.nextID.Load()
}

const legacySubscribeBuffer = 32

// Subscribe is retained for in-process callers (tests and anything that
// wants the raw event stream without going through HTTP).
func (h *SSEHub) Subscribe() (<-chan string, func()) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "realtime.SSEHub.Subscribe")
	sub := &Subscriber{
		legacy:    make(chan string, legacySubscribeBuffer),
		Cancelled: make(chan struct{}),
	}
	h.mu.Lock()
	h.subs[sub] = struct{}{}
	n := len(h.subs)
	h.mu.Unlock()
	middleware.RecordSSESubscriberGauge(n)
	cancel := func() {
		h.mu.Lock()
		_, ok := h.subs[sub]
		if ok {
			delete(h.subs, sub)
		}
		n := len(h.subs)
		h.mu.Unlock()
		middleware.RecordSSESubscriberGauge(n)
	}
	return sub.legacy, cancel
}

var _ Publisher = (*SSEHub)(nil)

// HeartbeatPeriod returns the configured SSE comment heartbeat interval.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *SSEHub) HeartbeatPeriod() time.Duration {
	if h == nil {
		return 0
	}
	return h.heartbeatPeriod
}
