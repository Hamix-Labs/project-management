// Package handler registers /draft-assist* HTTP routes for taskapi.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

const heartbeatInterval = 3 * time.Second

// Deps wires draft-assist HTTP handlers.
type Deps struct {
	Store  contract.Store
	Runner contract.Runner
}

// Handler serves draft-assist routes.
type Handler struct {
	store  contract.Store
	runner contract.Runner
}

// Register mounts draft-assist routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	h := &Handler{store: deps.Store, runner: deps.Runner}
	m.Handle("GET /draft-assist/ready", http.HandlerFunc(h.ready))
	m.Handle("POST /draft-assist/sessions", http.HandlerFunc(h.createSession))
	m.Handle("PUT /draft-assist/sessions/{id}/snapshot", http.HandlerFunc(h.putSnapshot))
	m.Handle("GET /draft-assist/sessions/{id}/events", http.HandlerFunc(h.events))
	m.Handle("POST /draft-assist/sessions/{id}/runs", http.HandlerFunc(h.startRun))
	m.Handle("POST /draft-assist/sessions/{id}/runs/{runId}/cancel", http.HandlerFunc(h.cancelRun))
	m.Handle("DELETE /draft-assist/sessions/{id}", http.HandlerFunc(h.deleteSession))
	m.Handle("GET /draft-assist/sessions/{id}", http.HandlerFunc(h.getSession))
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.ready")
	op := "draftAssist.ready"
	name := "missing"
	ready := false
	reason := "no runner configured"
	if h.runner != nil {
		name = h.runner.Name()
		ready = true
		reason = ""
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, map[string]any{
		"ready":  ready,
		"runner": name,
		"reason": reason,
	})
}

type createSessionBody struct {
	WorktreeID string              `json:"worktree_id"`
	Snapshot   domain.FormSnapshot `json:"snapshot"`
}

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.createSession")
	op := "draftAssist.createSession"
	var body createSessionBody
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "invalid request body")
		return
	}
	sess, err := h.store.CreateSession(r.Context(), contract.CreateSessionInput{
		WorktreeID: strings.TrimSpace(body.WorktreeID),
		Snapshot:   body.Snapshot,
	})
	if err != nil {
		writeStoreErr(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusCreated, map[string]any{
		"id":          sess.ID,
		"nonce":       sess.Nonce,
		"worktree_id": sess.WorktreeID,
		"snapshot":    sess.Snapshot,
		"created_at":  sess.CreatedAt,
	})
}

func (h *Handler) getSession(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.getSession")
	op := "draftAssist.getSession"
	id := r.PathValue("id")
	sess, err := h.store.GetSession(r.Context(), id)
	if err != nil {
		writeStoreErr(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, map[string]any{
		"id":          sess.ID,
		"nonce":       sess.Nonce,
		"worktree_id": sess.WorktreeID,
		"snapshot":    sess.Snapshot,
		"created_at":  sess.CreatedAt,
		"updated_at":  sess.UpdatedAt,
	})
}

func (h *Handler) putSnapshot(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.putSnapshot")
	op := "draftAssist.putSnapshot"
	id := r.PathValue("id")
	var snap domain.FormSnapshot
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &snap); err != nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "invalid request body")
		return
	}
	sess, err := h.store.UpdateSnapshot(r.Context(), id, snap)
	if err != nil {
		writeStoreErr(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, map[string]any{
		"id":       sess.ID,
		"snapshot": sess.Snapshot,
	})
}

type startRunBody struct {
	UserMessage string              `json:"user_message"`
	Snapshot    domain.FormSnapshot `json:"snapshot"`
}

func (h *Handler) startRun(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.startRun")
	op := "draftAssist.startRun"
	id := r.PathValue("id")
	var body startRunBody
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "invalid request body")
		return
	}
	msg := strings.TrimSpace(body.UserMessage)
	if msg == "" {
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, "user_message is required")
		return
	}
	if body.Snapshot.Prompt != "" || body.Snapshot.Title != "" {
		if _, err := h.store.UpdateSnapshot(r.Context(), id, body.Snapshot); err != nil {
			writeStoreErr(w, r, op, err)
			return
		}
	}
	runID, err := h.store.StartRun(r.Context(), id, contract.RunInput{UserMessage: msg, Snapshot: body.Snapshot})
	if err != nil {
		writeStoreErr(w, r, op, err)
		return
	}
	// 202 immediately — runner work is async.
	handlerhttp.WriteJSON(w, r, op, http.StatusAccepted, map[string]any{"run_id": runID})

	if h.runner == nil {
		_ = h.store.Publish(context.Background(), id, domain.Event{
			Kind:  domain.EventError,
			RunID: runID,
			Data:  domain.ErrorEventData{Code: "no_runner", Message: "no draft-assist runner configured"},
		})
		_ = h.store.Publish(context.Background(), id, domain.Event{
			Kind:  domain.EventDone,
			RunID: runID,
			Data:  domain.DoneEventData{Status: domain.RunStatusFailed},
		})
		_ = h.store.FinishRun(context.Background(), id, runID)
		return
	}

	runCtx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		defer func() { _ = h.store.FinishRun(context.Background(), id, runID) }()
		handle := &emitHandle{store: h.store}
		in := contract.RunInput{UserMessage: msg}
		if sess, err := h.store.GetSession(runCtx, id); err == nil {
			in.Snapshot = sess.Snapshot
		}
		if err := h.runner.Run(runCtx, id, runID, in, handle); err != nil && !errors.Is(err, context.Canceled) {
			slog.Warn("draftassist runner failed", "cmd", calltrace.LogCmd, "operation", op, "err", err, "session_id", id, "run_id", runID)
			_ = handle.Emit(context.Background(), id, runID, domain.EventError, domain.ErrorEventData{
				Code:    "runner_error",
				Message: err.Error(),
			})
			_ = handle.Emit(context.Background(), id, runID, domain.EventDone, domain.DoneEventData{Status: domain.RunStatusFailed})
		}
	}()
	_ = cancel // cancel retained via runCtx; CancelRun cancels via store
	// Wire cancel: CancelRun calls store cancel. Re-bind by wrapping CancelRun path.
	// Store.CancelRun already invokes runCancel if set; StartRun created a cancel
	// that was discarded. Fix: store the cancel on the session via a second Start.
	// Simpler approach: listen for CancelRun by checking ctx in runner via a
	// dedicated cancel registry on the handler.
	h.trackCancel(id, runID, cancel)
}

type cancelKey struct{ session, run string }

var cancelRegistry = struct {
	mu sync.Mutex
	m  map[cancelKey]context.CancelFunc
}{m: map[cancelKey]context.CancelFunc{}}

func (h *Handler) trackCancel(sessionID, runID string, cancel context.CancelFunc) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.trackCancel")
	cancelRegistry.mu.Lock()
	defer cancelRegistry.mu.Unlock()
	cancelRegistry.m[cancelKey{sessionID, runID}] = cancel
}

func (h *Handler) cancelRun(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.cancelRun")
	op := "draftAssist.cancelRun"
	id := r.PathValue("id")
	runID := r.PathValue("runId")
	cancelRegistry.mu.Lock()
	cancel, ok := cancelRegistry.m[cancelKey{id, runID}]
	if ok {
		delete(cancelRegistry.m, cancelKey{id, runID})
	}
	cancelRegistry.mu.Unlock()
	if ok {
		cancel()
	}
	if err := h.store.CancelRun(r.Context(), id, runID); err != nil {
		writeStoreErr(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusAccepted, map[string]any{"run_id": runID, "status": "cancelling"})
}

func (h *Handler) deleteSession(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.deleteSession")
	op := "draftAssist.deleteSession"
	id := r.PathValue("id")
	if err := h.store.DeleteSession(r.Context(), id); err != nil {
		writeStoreErr(w, r, op, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) events(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.events")
	op := "draftAssist.events"
	id := r.PathValue("id")
	flusher, ok := w.(http.Flusher)
	if !ok {
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	sinceID := parseLastEventID(r.Header.Get("Last-Event-ID"))

	sess, err := h.store.GetSession(r.Context(), id)
	if err != nil {
		writeStoreErr(w, r, op, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Initial session event (not replayed from ring unless already published).
	if sinceID == 0 {
		ev := domain.Event{
			Kind: domain.EventSession,
			At:   time.Now().UTC(),
			Data: domain.SessionEventData{
				SessionID:  sess.ID,
				WorktreeID: sess.WorktreeID,
				Snapshot:   sess.Snapshot,
			},
		}
		_ = h.store.Publish(r.Context(), id, ev)
	}

	sub, replay, err := h.store.Subscribe(r.Context(), id, sinceID)
	if err != nil {
		return
	}
	defer sub.Cancel()

	for _, ev := range replay {
		if !writeSSEEvent(w, flusher, ev) {
			return
		}
	}

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			active, _ := h.store.RunActive(r.Context(), id)
			if active {
				if _, err := fmt.Fprintf(w, ": heartbeat\n\n"); err != nil {
					return
				}
				flusher.Flush()
			}
		case ev, ok := <-sub.Events:
			if !ok {
				return
			}
			if !writeSSEEvent(w, flusher, ev) {
				return
			}
		}
	}
}

type emitHandle struct {
	store contract.Store
}

//funclogmeasure:skip category=hot-path reason="Publish fan-out helper; runner emits the operation-level trace."
func (h *emitHandle) Emit(ctx context.Context, sessionID, runID string, kind domain.EventKind, data any) error {
	return h.store.Publish(ctx, sessionID, domain.Event{
		Kind:  kind,
		RunID: runID,
		At:    time.Now().UTC(),
		Data:  data,
	})
}

func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, ev domain.Event) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.writeSSEEvent")
	payload, err := json.Marshal(ev)
	if err != nil {
		return false
	}
	if _, err := fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", ev.ID, ev.Kind, payload); err != nil {
		return false
	}
	flusher.Flush()
	return true
}

func parseLastEventID(raw string) uint64 {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.parseLastEventID")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func writeStoreErr(w http.ResponseWriter, r *http.Request, op string, err error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.handler.writeStoreErr")
	switch {
	case errors.Is(err, domain.ErrNotFound):
		handlerhttp.WriteJSONError(w, r, op, http.StatusNotFound, "not found")
	case errors.Is(err, domain.ErrInvalidInput):
		handlerhttp.WriteJSONError(w, r, op, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrRunActive):
		handlerhttp.WriteJSONError(w, r, op, http.StatusConflict, "run already active")
	case errors.Is(err, domain.ErrNonceMismatch):
		handlerhttp.WriteJSONError(w, r, op, http.StatusForbidden, "nonce mismatch")
	default:
		handlerhttp.WriteJSONError(w, r, op, http.StatusInternalServerError, "internal error")
	}
}
