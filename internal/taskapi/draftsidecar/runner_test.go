package draftsidecar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
)

// captureHandle records events into a slice for assertions.
type captureHandle struct {
	mu     sync.Mutex
	events []captured
}

type captured struct {
	kind domain.EventKind
	data any
}

func (h *captureHandle) Emit(_ context.Context, _, _ string, kind domain.EventKind, data any) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, captured{kind: kind, data: data})
	return nil
}

func (h *captureHandle) len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.events)
}

// fakeSidecar returns an httptest.Server whose POST /runs streams SSE
// frames driven by a per-request emitter callback. POST /runs/{id}/cancel
// closes the current run's stream after emitting the terminal frames.
type fakeSidecar struct {
	*httptest.Server
	mu           sync.Mutex
	activeCancel func()
	runBody      map[string]any
}

func newFakeSidecar(t *testing.T, script func(w http.ResponseWriter, body map[string]any, cancelCh <-chan struct{})) *fakeSidecar {
	t.Helper()
	sc := &fakeSidecar{}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /runs", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sc.mu.Lock()
		sc.runBody = body
		cancelCh := make(chan struct{})
		sc.activeCancel = func() { close(cancelCh) }
		sc.mu.Unlock()

		w.Header().Set("content-type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		script(w, body, cancelCh)
	})
	mux.HandleFunc("POST /runs/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		sc.mu.Lock()
		cancel := sc.activeCancel
		sc.activeCancel = nil
		sc.mu.Unlock()
		if cancel != nil {
			cancel()
		}
		w.WriteHeader(http.StatusAccepted)
	})
	sc.Server = httptest.NewServer(mux)
	t.Cleanup(sc.Server.Close)
	return sc
}

func writeSSE(w http.ResponseWriter, id int, event string, payload any) error {
	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", id, event, buf); err != nil {
		return err
	}
	if fl, ok := w.(http.Flusher); ok {
		fl.Flush()
	}
	return nil
}

func TestRunner_HappyPath_EmitsAllFrames(t *testing.T) {
	sc := newFakeSidecar(t, func(w http.ResponseWriter, body map[string]any, cancelCh <-chan struct{}) {
		_ = writeSSE(w, 1, "session", domain.SessionEventData{SessionID: "sess"})
		_ = writeSSE(w, 2, "status", domain.StatusEventData{Status: domain.RunStatusThinking})
		_ = writeSSE(w, 3, "token", domain.TokenEventData{Delta: "hi"})
		_ = writeSSE(w, 4, "done", domain.DoneEventData{Status: domain.RunStatusDone})
	})

	r := NewRunner(RunnerOptions{BaseURLOverride: sc.URL})
	if r.Name() != "sdk" {
		t.Fatalf("name=%q want sdk", r.Name())
	}
	h := &captureHandle{}

	err := r.Run(context.Background(), "sess", "run", contract.RunInput{
		UserMessage: "tighten",
		WorktreeCwd: "/tmp/x",
	}, h)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if h.len() != 4 {
		t.Fatalf("events=%d want 4", h.len())
	}
	if h.events[0].kind != domain.EventSession {
		t.Fatalf("first event kind=%v want session", h.events[0].kind)
	}
	if last := h.events[len(h.events)-1]; last.kind != domain.EventDone {
		t.Fatalf("last event kind=%v want done", last.kind)
	}
}

func TestRunner_CancelForwardsAndDrainsCancelled(t *testing.T) {
	streamStarted := make(chan struct{})
	sc := newFakeSidecar(t, func(w http.ResponseWriter, body map[string]any, cancelCh <-chan struct{}) {
		_ = writeSSE(w, 1, "session", domain.SessionEventData{SessionID: "sess"})
		close(streamStarted)
		_ = writeSSE(w, 2, "status", domain.StatusEventData{Status: domain.RunStatusThinking})
		select {
		case <-cancelCh:
		case <-time.After(3 * time.Second):
		}
		_ = writeSSE(w, 3, "status", domain.StatusEventData{Status: domain.RunStatusCancelling})
		_ = writeSSE(w, 4, "done", domain.DoneEventData{Status: domain.RunStatusCancelled})
	})

	r := NewRunner(RunnerOptions{BaseURLOverride: sc.URL})
	h := &captureHandle{}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- r.Run(ctx, "sess", "run-x", contract.RunInput{
			UserMessage: "tighten",
			WorktreeCwd: "/tmp/x",
		}, h)
	}()

	select {
	case <-streamStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("stream never started")
	}

	cancel()

	select {
	case err := <-done:
		if err != nil && err != context.Canceled {
			t.Fatalf("Run returned %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after cancel")
	}

	// The cancelled done frame may or may not arrive before context tear
	// down; we just require at least session + status.
	if h.len() < 2 {
		t.Fatalf("events=%d want ≥2", h.len())
	}
}

func TestRunner_HTTPErrorSurfacedAsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	t.Cleanup(srv.Close)

	r := NewRunner(RunnerOptions{BaseURLOverride: srv.URL})
	err := r.Run(context.Background(), "sess", "run", contract.RunInput{
		UserMessage: "hi",
		WorktreeCwd: "/tmp/x",
	}, &captureHandle{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunner_DisconnectDuringStream_ReturnsErrOrNil(t *testing.T) {
	sc := newFakeSidecar(t, func(w http.ResponseWriter, body map[string]any, cancelCh <-chan struct{}) {
		_ = writeSSE(w, 1, "session", domain.SessionEventData{SessionID: "sess"})
		// Server closes the response without a done frame. Runner
		// should either return nil (EOF) or the current ctx err.
	})

	r := NewRunner(RunnerOptions{BaseURLOverride: sc.URL})
	h := &captureHandle{}
	err := r.Run(context.Background(), "sess", "run", contract.RunInput{
		UserMessage: "hi",
		WorktreeCwd: "/tmp/x",
	}, h)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if h.len() < 1 {
		t.Fatalf("events=%d want ≥1", h.len())
	}
}

func TestRunner_PortSource_WhenPortZero_Errors(t *testing.T) {
	r := NewRunner(RunnerOptions{PortSource: portZero{}})
	err := r.Run(context.Background(), "sess", "run", contract.RunInput{
		UserMessage: "hi",
		WorktreeCwd: "/tmp/x",
	}, &captureHandle{})
	if err == nil {
		t.Fatal("expected error when port not known")
	}
}

type portZero struct{}

func (portZero) Port() int { return 0 }
