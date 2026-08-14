package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"log/slog"
)

const (
	// eventRingCap bounds Last-Event-ID replay. Documented in
	// docs/domain/draft-assist.md; bump only with an ADR note.
	eventRingCap     = 256
	subscriberBufCap = 64
)

type subscriber struct {
	ch chan domain.Event
}

type sessionState struct {
	sess      domain.Session
	activeRun string
	runCancel context.CancelFunc
	seq       uint64
	ring      []domain.Event
	subs      map[uint64]*subscriber
	nextSub   uint64
}

// MemoryStore is the in-memory session + event bus.
type MemoryStore struct {
	mu       sync.Mutex
	sessions map[string]*sessionState
}

// NewMemoryStore constructs an empty store.
//
//funclogmeasure:skip category=hot-path reason="Constructor; mutating methods emit operation traces."
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: make(map[string]*sessionState)}
}

//funclogmeasure:skip category=hot-path reason="Pure id helper without I/O beyond crypto/rand."
func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func (s *MemoryStore) CreateSession(_ context.Context, in contract.CreateSessionInput) (*domain.Session, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.MemoryStore.CreateSession")
	now := time.Now().UTC()
	snap := in.Snapshot
	snap.UpdatedAt = now
	id := newID()
	nonce := newID()
	sess := domain.Session{
		ID:         id,
		Nonce:      nonce,
		WorktreeID: in.WorktreeID,
		Snapshot:   snap,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	st := &sessionState{
		sess: sess,
		subs: make(map[uint64]*subscriber),
	}
	s.mu.Lock()
	s.sessions[id] = st
	s.mu.Unlock()
	copy := sess
	return &copy, nil
}

func (s *MemoryStore) GetSession(_ context.Context, id string) (*domain.Session, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.MemoryStore.GetSession", "session_id", id)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copy := st.sess
	return &copy, nil
}

func (s *MemoryStore) UpdateSnapshot(_ context.Context, id string, snap domain.FormSnapshot) (*domain.Session, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.MemoryStore.UpdateSnapshot", "session_id", id)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	now := time.Now().UTC()
	snap.UpdatedAt = now
	st.sess.Snapshot = snap
	st.sess.UpdatedAt = now
	copy := st.sess
	return &copy, nil
}

func (s *MemoryStore) UpdatePrompt(_ context.Context, id, nonce, prompt string) (*domain.Session, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.MemoryStore.UpdatePrompt", "session_id", id)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if st.sess.Nonce != nonce {
		return nil, domain.ErrNonceMismatch
	}
	now := time.Now().UTC()
	st.sess.Snapshot.Prompt = prompt
	st.sess.Snapshot.UpdatedAt = now
	st.sess.UpdatedAt = now
	copy := st.sess
	return &copy, nil
}

func (s *MemoryStore) DeleteSession(_ context.Context, id string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.MemoryStore.DeleteSession", "session_id", id)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[id]
	if !ok {
		return domain.ErrNotFound
	}
	if st.runCancel != nil {
		st.runCancel()
	}
	for _, sub := range st.subs {
		close(sub.ch)
	}
	delete(s.sessions, id)
	return nil
}

func (s *MemoryStore) StartRun(ctx context.Context, id string, in contract.RunInput) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.MemoryStore.StartRun", "session_id", id)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[id]
	if !ok {
		return "", domain.ErrNotFound
	}
	if st.activeRun != "" {
		return "", domain.ErrRunActive
	}
	if in.UserMessage == "" {
		return "", fmt.Errorf("%w: empty user message", domain.ErrInvalidInput)
	}
	runID := newID()
	runCtx, cancel := context.WithCancel(ctx)
	st.activeRun = runID
	st.runCancel = cancel
	_ = runCtx
	in.Snapshot = st.sess.Snapshot
	return runID, nil
}

// BindRunContext stores a cancel func derived from StartRun's parent; called by handler.
//
//funclogmeasure:skip category=hot-path reason="Internal cancel wiring; CreateRun emits the operation trace."
func (s *MemoryStore) BindRunCancel(sessionID, runID string, cancel context.CancelFunc) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[sessionID]
	if !ok {
		return domain.ErrNotFound
	}
	if st.activeRun != runID {
		return domain.ErrNotFound
	}
	st.runCancel = cancel
	return nil
}

func (s *MemoryStore) CancelRun(_ context.Context, sessionID, runID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.MemoryStore.CancelRun", "session_id", sessionID, "run_id", runID)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[sessionID]
	if !ok {
		return domain.ErrNotFound
	}
	if st.activeRun != runID {
		return domain.ErrNotFound
	}
	if st.runCancel != nil {
		st.runCancel()
	}
	return nil
}

func (s *MemoryStore) FinishRun(_ context.Context, sessionID, runID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.MemoryStore.FinishRun", "session_id", sessionID, "run_id", runID)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[sessionID]
	if !ok {
		return domain.ErrNotFound
	}
	if st.activeRun != runID {
		return nil
	}
	st.activeRun = ""
	st.runCancel = nil
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure accessor; CreateRun/CompleteRun emit operation traces."
func (s *MemoryStore) RunActive(_ context.Context, sessionID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[sessionID]
	if !ok {
		return false, domain.ErrNotFound
	}
	return st.activeRun != "", nil
}

//funclogmeasure:skip category=hot-path reason="Pure accessor; CreateRun/CompleteRun emit operation traces."
func (s *MemoryStore) ActiveRunID(sessionID string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[sessionID]
	if !ok {
		return "", false
	}
	return st.activeRun, st.activeRun != ""
}

func (s *MemoryStore) Publish(_ context.Context, sessionID string, ev domain.Event) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.MemoryStore.Publish", "session_id", sessionID, "kind", string(ev.Kind))
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[sessionID]
	if !ok {
		return domain.ErrNotFound
	}
	st.seq++
	ev.ID = st.seq
	if ev.At.IsZero() {
		ev.At = time.Now().UTC()
	}
	st.ring = append(st.ring, ev)
	if len(st.ring) > eventRingCap {
		st.ring = st.ring[len(st.ring)-eventRingCap:]
	}
	for id, sub := range st.subs {
		select {
		case sub.ch <- ev:
		default:
			// Drop slow subscriber to protect the bus; SPA reconnects via Last-Event-ID.
			close(sub.ch)
			delete(st.subs, id)
		}
	}
	return nil
}

func (s *MemoryStore) Subscribe(_ context.Context, sessionID string, sinceID uint64) (*contract.Subscription, []domain.Event, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "draftassist.MemoryStore.Subscribe", "session_id", sessionID, "since_id", sinceID)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sessions[sessionID]
	if !ok {
		return nil, nil, domain.ErrNotFound
	}
	var replay []domain.Event
	for _, ev := range st.ring {
		if ev.ID > sinceID {
			replay = append(replay, ev)
		}
	}
	ch := make(chan domain.Event, subscriberBufCap)
	st.nextSub++
	subID := st.nextSub
	st.subs[subID] = &subscriber{ch: ch}
	cancel := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		st2, ok := s.sessions[sessionID]
		if !ok {
			return
		}
		if sub, ok := st2.subs[subID]; ok {
			delete(st2.subs, subID)
			close(sub.ch)
		}
	}
	return &contract.Subscription{Events: ch, Cancel: cancel}, replay, nil
}

// Ensure MemoryStore satisfies contract.Store.
var _ contract.Store = (*MemoryStore)(nil)
