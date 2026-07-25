// Package store implements GORM persistence for task CRUD, dependencies,
// ready queue, stats, dev-mirror row updates, and health probes.
package store

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/internal/devmirror"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/internal/health"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/internal/ready"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/internal/stats"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/internal/tasks"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"gorm.io/gorm"
)

// Store is the GORM-backed persistence facade for taskcore concerns.
type Store struct {
	db *gorm.DB
}

// NewStore returns a Store backed by db.
func NewStore(db *gorm.DB) *Store {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.NewStore")
	return &Store{db: db}
}

// DB exposes the underlying GORM handle for composition wiring in internal/taskapi/composition.
//
//funclogmeasure:skip category=hot-path reason="Test-only accessor; no store operation boundary."
func (s *Store) DB() *gorm.DB {
	if s == nil {
		return nil
	}
	return s.db
}

// DefaultReadyTimeout is the recommended upper bound for readiness probes.
const DefaultReadyTimeout = 2 * time.Second

// FailedPredicate identifies the first worker readiness check that failed.
type FailedPredicate = taskcorecontract.FailedPredicate

type (
	CreateTaskInput         = tasks.CreateInput
	UpdateTaskInput         = tasks.UpdateInput
	ProjectFieldPatch       = tasks.ProjectFieldPatch
	PickupNotBeforePatch    = tasks.PickupNotBeforePatch
	RequestRetryInput       = tasks.RequestRetryInput
	RequestPolishInput      = tasks.RequestPolishInput
	AgentPickupResult       = tasks.AgentPickupResult
	ListFilter              = tasks.ListFilter
	ReadyTaskQueueCursor    = ready.QueueCursor
	ReadyTaskQueueCandidate = ready.QueueCandidate
	DeferredPickup          = ready.DeferredPickup
	DeferredPickupCursor    = ready.DeferredPickupCursor
	TaskStats               = stats.TaskStats
	CycleStats              = stats.CycleStats
	PhaseStats              = stats.PhaseStats
	RunnerStats             = stats.RunnerStats
	RunnerBucket            = stats.RunnerBucket
	RecentFailure           = stats.RecentFailure
	PreFeatureCycleCounts   = stats.PreFeatureCycleCounts
)

const (
	RunnerUnknownKey = stats.RunnerUnknownKey
)

// ShouldNotifyReadyNow returns true when a freshly-ready task should enter the in-memory queue.
func ShouldNotifyReadyNow(pickupNotBefore *time.Time, now time.Time) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.taskcorestore.ShouldNotifyReadyNow", "has_pickup", pickupNotBefore != nil)
	return taskcorecontract.ShouldNotifyReadyNow(pickupNotBefore, now)
}

func (s *Store) Get(ctx context.Context, id string) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.Get")
	return tasks.Get(ctx, s.db, id)
}

func (s *Store) AgentPickup(ctx context.Context, taskID string, by domain.Actor) (*AgentPickupResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.AgentPickup", "task_id", taskID)
	return tasks.AgentPickup(ctx, s.db, taskID, by)
}

func (s *Store) RequestTaskRetry(ctx context.Context, in RequestRetryInput, by domain.Actor) (*domain.Task, domain.Status, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.RequestTaskRetry", "task_id", in.TaskID)
	return tasks.RequestTaskRetry(ctx, s.db, in, by)
}

func (s *Store) RequestTaskApprove(ctx context.Context, taskID string, by domain.Actor) (*domain.Task, domain.Status, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.RequestTaskApprove", "task_id", taskID)
	return tasks.RequestTaskApprove(ctx, s.db, taskID, by)
}

func (s *Store) RequestTaskPolish(ctx context.Context, in RequestPolishInput, by domain.Actor) (*domain.Task, domain.Status, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.RequestTaskPolish", "task_id", in.TaskID)
	return tasks.RequestTaskPolish(ctx, s.db, in, by)
}

func (s *Store) Create(ctx context.Context, in CreateTaskInput, by domain.Actor) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.Create")
	return tasks.Create(ctx, s.db, in, by)
}

func (s *Store) Update(ctx context.Context, id string, in UpdateTaskInput, by domain.Actor) (*domain.Task, domain.Status, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.Update")
	return tasks.Update(ctx, s.db, id, in, by)
}

func (s *Store) Delete(ctx context.Context, id string, by domain.Actor) ([]string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.Delete")
	return tasks.Delete(ctx, s.db, id, by)
}

func (s *Store) Close(ctx context.Context, id string, by domain.Actor) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.Close", "task_id", id)
	return tasks.Close(ctx, s.db, id, by)
}

func (s *Store) Reopen(ctx context.Context, id string, by domain.Actor) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.Reopen", "task_id", id)
	return tasks.Reopen(ctx, s.db, id, by)
}

func (s *Store) ListFlat(ctx context.Context, limit, offset int, filter *ListFilter) ([]domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ListFlat")
	return tasks.ListFlat(ctx, s.db, limit, offset, filter)
}

func (s *Store) ListFlatPage(ctx context.Context, limit, offset int, filter *ListFilter) ([]domain.Task, bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ListFlatPage")
	return tasks.ListFlatPage(ctx, s.db, limit, offset, filter)
}

func (s *Store) ListFlatAfter(ctx context.Context, limit int, afterID string) ([]domain.Task, bool, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ListFlatAfter")
	return tasks.ListFlatAfter(ctx, s.db, limit, afterID)
}

func (s *Store) AddTaskDependency(ctx context.Context, taskID, dependsOnTaskID string, satisfies domain.DependencySatisfies) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.AddTaskDependency")
	return tasks.AddDependency(ctx, s.db, taskID, dependsOnTaskID, satisfies)
}

func (s *Store) RemoveTaskDependency(ctx context.Context, taskID, dependsOnTaskID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.RemoveTaskDependency")
	return tasks.RemoveDependency(ctx, s.db, taskID, dependsOnTaskID)
}

func (s *Store) ListTaskDependencies(ctx context.Context, taskID string) ([]domain.DependencyEdge, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ListTaskDependencies")
	return tasks.ListDependencyEdges(ctx, s.db, taskID)
}

func (s *Store) SetTaskDependencies(ctx context.Context, taskID string, dependsOn []domain.DependencyEdge) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.SetTaskDependencies")
	return tasks.SetDependencies(ctx, s.db, taskID, dependsOn)
}

func (s *Store) ListDependents(ctx context.Context, predecessorID string) ([]string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ListDependents")
	return tasks.ListDependents(ctx, s.db, predecessorID)
}

func (s *Store) ReadyForAgentPickup(ctx context.Context, t *domain.Task, now time.Time) (bool, FailedPredicate, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ReadyForAgentPickup")
	return tasks.ReadyForAgentPickup(ctx, s.db, t, now)
}

func (s *Store) ApplyTaskGateAction(ctx context.Context, taskID string, action taskcorecontract.GateAction, by domain.Actor) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ApplyTaskGateAction")
	return tasks.ApplyTaskGateAction(ctx, s.db, taskID, string(action), by)
}

func (s *Store) ListDeferredReadyPickupTasks(ctx context.Context, limit int, after *DeferredPickupCursor) ([]DeferredPickup, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ListDeferredReadyPickupTasks")
	return ready.ListDeferredReadyPickups(ctx, s.db, time.Now().UTC(), limit, after)
}

func (s *Store) ListReadyTaskQueueCandidates(ctx context.Context, limit int, cursor *ReadyTaskQueueCursor) ([]ReadyTaskQueueCandidate, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ListReadyTaskQueueCandidates")
	return ready.ListQueueCandidates(ctx, s.db, limit, cursor)
}

func (s *Store) ListReadyTasksUserCreated(ctx context.Context, limit int, afterID string) ([]domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ListReadyTasksUserCreated")
	return ready.ListUserCreated(ctx, s.db, limit, afterID)
}

// PendingWorktreeRow is a ready task awaiting managed-worktree allocate.
type PendingWorktreeRow = ready.PendingWorktreeRow

// ListTasksPendingWorktree lists ready tasks with no worktree_id (ADR-0083).
func (s *Store) ListTasksPendingWorktree(ctx context.Context, limit int) ([]PendingWorktreeRow, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ListTasksPendingWorktree")
	return ready.ListPendingWorktree(ctx, s.db, limit)
}

func (s *Store) TaskStats(ctx context.Context) (TaskStats, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.TaskStats")
	return stats.Get(ctx, s.db)
}

func (s *Store) CountPreFeatureCycles(ctx context.Context) (PreFeatureCycleCounts, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.CountPreFeatureCycles")
	return stats.CountPreFeatureCycles(ctx, s.db)
}

// CountTasksByWorktreeID returns how many tasks still reference worktreeID.
func (s *Store) CountTasksByWorktreeID(ctx context.Context, worktreeID string) (int64, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.CountTasksByWorktreeID")
	worktreeID = strings.TrimSpace(worktreeID)
	if worktreeID == "" {
		return 0, nil
	}
	var n int64
	err := s.db.WithContext(ctx).Model(&model.Task{}).Where("worktree_id = ?", worktreeID).Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("count tasks by worktree_id: %w", err)
	}
	return n, nil
}

func (s *Store) ApplyDevTaskRowMirror(ctx context.Context, taskID string, typ taskeventsdomain.EventType, data []byte) (*domain.Task, domain.Status, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ApplyDevTaskRowMirror")
	return devmirror.ApplyTaskRowMirror(ctx, s.db, taskID, typ, data)
}

func (s *Store) ListDevsimTasks(ctx context.Context, idLikePattern string) ([]domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.ListDevsimTasks")
	return devmirror.ListDevsimTasks(ctx, s.db, idLikePattern)
}

func (s *Store) Ping(ctx context.Context) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.Ping")
	if s == nil {
		return health.Ping(ctx, nil)
	}
	return health.Ping(ctx, s.db)
}

func (s *Store) Ready(ctx context.Context) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcore.store.Ready")
	if s == nil {
		return health.Ready(ctx, nil)
	}
	return health.Ready(ctx, s.db)
}
