package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"log/slog"
	"sync"
	"time"

	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	projectsstore "github.com/AlexsanderHamir/Hamix/pkgs/projects/store"
	settingsstore "github.com/AlexsanderHamir/Hamix/pkgs/settings/store"
	checkliststore "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store"
	composestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/store"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	taskeventsstore "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/internal/notify"
	"gorm.io/gorm"
)

// Store is the public GORM-backed persistence facade for tasks, audit
// events, checklists, drafts, evaluations, cycles/phases, the ready-task
// queue, dev-mirror, and health probes. Behavior is split across
// internal/<domain>/ subpackages; the methods on *Store delegate. See
// pkgs/tasks/store/README.md for the concern map.
type Store struct {
	db        *gorm.DB
	taskcore  *taskcorestore.Store
	projects  *projectsstore.Store
	git       *gitinventorystore.Store
	settings  *settingsstore.Store
	compose   *composestore.Store
	checklist *checkliststore.Store
	cycles    *cyclesstore.Store
	events    *taskeventsstore.Store
	notify    notify.Holder

	pickupWakeMu sync.RWMutex
	pickupWake   PickupWake
}

// NewStore returns a Store backed by db. The caller still configures
// ready-task notifications via SetReadyTaskNotifier after construction.
func NewStore(db *gorm.DB) *Store {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.NewStore")
	return &Store{
		db:        db,
		taskcore:  taskcorestore.NewStore(db),
		projects:  projectsstore.NewStore(db),
		git:       gitinventorystore.NewStore(db),
		settings:  settingsstore.NewStore(db),
		compose:   composestore.NewStore(db),
		checklist: checkliststore.NewStore(db),
		cycles:    cyclesstore.NewStore(db),
		events:    taskeventsstore.NewStore(db),
	}
}

// ReadyTaskNotifier is invoked by the store after a task row is committed with status ready
// (on create) or when status transitions to ready (on update or dev row mirror). The store may
// hold a nil notifier (for example in tests); taskapi wires a non-nil implementation at startup.
// Implementations should avoid blocking the store caller for long (for example use a buffered channel).
//
// The interface lives in pkgs/tasks/store/internal/notify so subpackages can publish without
// taking a dependency on the public facade. The alias here keeps existing callers
// (cmd/taskapi/run_helpers.go, tests) compiling unchanged.
type ReadyTaskNotifier = notify.Notifier

// SetReadyTaskNotifier registers n for ready-task notifications. Pass nil to clear the notifier.
// Safe for use before serving traffic; typical wiring is once at process startup.
func (s *Store) SetReadyTaskNotifier(n ReadyTaskNotifier) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SetReadyTaskNotifier", "enabled", n != nil)
	if s == nil {
		return
	}
	s.notify.Set(n)
}

// SetPickupWake registers w for deferred-pickup scheduling (nil clears).
// Typical wiring: once at taskapi startup after SetReadyTaskNotifier.
func (s *Store) SetPickupWake(w PickupWake) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.SetPickupWake", "enabled", w != nil)
	if s == nil {
		return
	}
	s.pickupWakeMu.Lock()
	defer s.pickupWakeMu.Unlock()
	s.pickupWake = w
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Store) schedulePickupWake(ctx context.Context, taskID string, notBefore time.Time) {
	s.pickupWakeMu.RLock()
	w := s.pickupWake
	s.pickupWakeMu.RUnlock()
	if w == nil || taskID == "" {
		return
	}
	w.Schedule(ctx, taskID, notBefore)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Store) cancelPickupWake(taskID string) {
	s.pickupWakeMu.RLock()
	w := s.pickupWake
	s.pickupWakeMu.RUnlock()
	if w == nil || taskID == "" {
		return
	}
	w.Cancel(taskID)
}

// notifyReadyTask is the package-internal entrypoint used by CRUD,
// update, and dev-mirror code paths. It forwards to the holder so the
// concurrency policy lives in one place.
func (s *Store) notifyReadyTask(ctx context.Context, task taskcoredomain.Task) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.notifyReadyTask", "task_id", task.ID)
	if s == nil {
		return
	}
	s.notify.Notify(ctx, task)
}
