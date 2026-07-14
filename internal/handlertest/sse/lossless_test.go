package sse_test

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/middleware"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestHTTP_SSE_emitsIDLineForEventSourceResume(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	streamReady := make(chan struct{})
	gotID := make(chan uint64, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/events", nil)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer res.Body.Close()

		br := bufio.NewReader(res.Body)
		_, _ = br.ReadString('\n') // retry: 3000
		_, _ = br.ReadString('\n') // blank line
		close(streamReady)

		for {
			line, err := br.ReadString('\n')
			if err != nil {
				return
			}
			s := strings.TrimSpace(line)
			if !strings.HasPrefix(s, "id:") {
				continue
			}
			idStr := strings.TrimSpace(strings.TrimPrefix(s, "id:"))
			n, perr := strconv.ParseUint(idStr, 10, 64)
			if perr != nil {
				t.Errorf("invalid id line %q", s)
				return
			}
			gotID <- n
			return
		}
	}()

	<-streamReady
	res, err := http.Post(srv.URL+"/tasks", "application/json", strings.NewReader(handlertest.WithCreateChecklistForURL(srv.URL, `{"title":"sse","priority":"medium"}`)))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()

	select {
	case id := <-gotID:
		if id == 0 {
			t.Fatalf("event id must be > 0 (monotonic counter starts at 1), got %d", id)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for id: line on the wire")
	}
}

// TestHTTP_SSE_replaysOnReconnectWithLastEventID is the end-to-end
// proof that the wire contract works as designed. We publish 3
// events into the hub before any client connects, then connect with
// `Last-Event-ID: 0` and assert the replay tail arrives in order.
// Without this test a future refactor that wired Last-Event-ID
// only on the in-memory subscribe path (skipping the HTTP header
// parser) would silently break browser reconnects.
func TestHTTP_SSE_replaysOnReconnectWithLastEventID(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	s := composition.NewAPI(db)
	hub := realtime.NewSSEHubWith(realtime.SSEHubOptions{RingSize: 16, SubscriberBuffer: 32})
	h := handler.NewHandler(s, hub, nil)
	srv := httptest.NewServer(h)
	defer srv.Close()

	hub.Publish(realtime.Event{Type: realtime.TaskUpdated, ID: "first"})  // id=1
	hub.Publish(realtime.Event{Type: realtime.TaskUpdated, ID: "second"}) // id=2
	hub.Publish(realtime.Event{Type: realtime.TaskUpdated, ID: "third"})  // id=3

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/events", nil)
	// Simulate an EventSource reconnect after seeing id=1 — the
	// browser would set Last-Event-ID to the last id it captured.
	// We expect ids 2 and 3 to replay before the live tail loop.
	req.Header.Set("Last-Event-ID", "1")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	br := bufio.NewReader(res.Body)
	var ids []string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && len(ids) < 2 {
		line, err := br.ReadString('\n')
		if err != nil {
			break
		}
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "data:") && strings.Contains(s, "task_updated") {
			ids = append(ids, s)
		}
	}
	if len(ids) != 2 {
		t.Fatalf("replay: got %d events, want 2 (lines=%v)", len(ids), ids)
	}
	if !strings.Contains(ids[0], `"id":"second"`) ||
		!strings.Contains(ids[1], `"id":"third"`) {
		t.Fatalf("replay order wrong (want second,third): %v", ids)
	}
}

// TestHTTP_SSE_emitsResyncWhenLastEventIDOutsideRing verifies the
// gap-on-reconnect path: a client whose Last-Event-ID is older than
// the oldest retained ring entry receives one `data: {"type":"resync"}`
// directive. The SPA's useTaskEventStream handler then drops every
// cache and refetches from REST — the documented escape hatch when
// the in-memory ring can't bridge the gap.
func TestHTTP_SSE_emitsResyncWhenLastEventIDOutsideRing(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	s := composition.NewAPI(db)
	// 4-entry ring + 6 publishes → ids 1..2 are evicted, oldest
	// retained id = 3.
	hub := realtime.NewSSEHubWith(realtime.SSEHubOptions{RingSize: 4, SubscriberBuffer: 32})
	h := handler.NewHandler(s, hub, nil)
	srv := httptest.NewServer(h)
	defer srv.Close()

	for i := 1; i <= 6; i++ {
		hub.Publish(realtime.Event{Type: realtime.TaskUpdated, ID: fmt.Sprintf("t-%d", i)})
	}

	resyncBefore := testutil.ToFloat64(middleware.SSEResyncEmittedCounter())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/events", nil)
	req.Header.Set("Last-Event-ID", "1") // outside the 4-entry window
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	br := bufio.NewReader(res.Body)
	sawResync := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line, err := br.ReadString('\n')
		if err != nil {
			break
		}
		if strings.Contains(line, `"type":"resync"`) {
			sawResync = true
			break
		}
	}
	if !sawResync {
		t.Fatalf("expected one resync directive on the wire after gap reconnect")
	}
	if got, want := testutil.ToFloat64(middleware.SSEResyncEmittedCounter()), resyncBefore+1; got < want {
		t.Fatalf("resync counter=%v want >=%v", got, want)
	}
}

// TestHTTP_SSE_heartbeatLineKeepsConnectionAlive verifies that a
// silent stream still gets `: heartbeat` comment lines on the
// configured cadence. Browsers ignore the comment line per the SSE
// spec, but reverse proxies (and corporate VPN gateways) typically
// idle-kill TCP connections after 30-60s with no traffic — without
// the heartbeat the client would see a forced disconnect every
// minute even when the server is healthy.
func TestHTTP_SSE_heartbeatLineKeepsConnectionAlive(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	hub := realtime.NewSSEHubWith(realtime.SSEHubOptions{
		RingSize:         16,
		SubscriberBuffer: 32,
		HeartbeatPeriod:  50 * time.Millisecond,
	})
	h := handler.NewHandler(composition.NewAPI(db), hub, nil)
	srv := httptest.NewServer(h)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/events", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	br := bufio.NewReader(res.Body)
	saw := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line, err := br.ReadString('\n')
		if err != nil {
			break
		}
		if strings.HasPrefix(strings.TrimSpace(line), ":") {
			saw = true
			break
		}
	}
	if !saw {
		t.Fatalf("expected at least one `: heartbeat` line within 2s (period=50ms)")
	}
}
