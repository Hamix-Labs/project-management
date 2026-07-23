package composition

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/scheduling"
)

const worktreeProvisionReconcileInterval = 30 * time.Second

// WorktreeProvisioner allocates managed worktrees after task create (ADR-0083).
type WorktreeProvisioner struct {
	api  *API
	hub  *realtime.SSEHub
	jobs chan provisionJob

	mu       sync.Mutex
	inflight map[string]struct{}

	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}
}

type provisionJob struct {
	taskID       string
	repositoryID string
}

// NewWorktreeProvisioner returns a provisioner that must be started with Run.
//
//funclogmeasure:skip category=hot-path reason="Composition wiring; Run emits operation traces."
func NewWorktreeProvisioner(api *API, hub *realtime.SSEHub) *WorktreeProvisioner {
	return &WorktreeProvisioner{
		api:      api,
		hub:      hub,
		jobs:     make(chan provisionJob, 64),
		inflight: make(map[string]struct{}),
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Enqueue schedules allocate for taskID (deduped while in-flight).
func (p *WorktreeProvisioner) Enqueue(ctx context.Context, taskID, repositoryID string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "composition.WorktreeProvisioner.Enqueue",
		"task_id", taskID, "repository_id", repositoryID)
	taskID = strings.TrimSpace(taskID)
	repositoryID = strings.TrimSpace(repositoryID)
	if p == nil || taskID == "" || repositoryID == "" {
		return
	}
	select {
	case <-p.stopCh:
		return
	default:
	}
	job := provisionJob{taskID: taskID, repositoryID: repositoryID}
	select {
	case p.jobs <- job:
	default:
		slog.Warn("worktree provision queue full; relying on reconcile",
			"task_id", taskID, "repository_id", repositoryID)
	}
}

// Stop signals the Run loop to exit.
func (p *WorktreeProvisioner) Stop() {
	if p == nil {
		return
	}
	p.stopOnce.Do(func() { close(p.stopCh) })
	select {
	case <-p.doneCh:
	case <-time.After(15 * time.Second):
		slog.Warn("worktree provisioner stop timed out")
	}
}

// Run processes enqueue + periodic reconcile until Stop or ctx cancel.
func (p *WorktreeProvisioner) Run(ctx context.Context) {
	defer close(p.doneCh)
	if p == nil {
		return
	}
	slog.Info("worktree provisioner started", "cmd", calltrace.LogCmd,
		"operation", "composition.WorktreeProvisioner.Run",
		"reconcile_interval", worktreeProvisionReconcileInterval.String())
	p.reconcile(ctx)
	ticker := time.NewTicker(worktreeProvisionReconcileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case job := <-p.jobs:
			p.provisionOne(ctx, job.taskID, job.repositoryID)
		case <-ticker.C:
			p.reconcile(ctx)
		}
	}
}

func (p *WorktreeProvisioner) reconcile(ctx context.Context) {
	if p == nil || p.api == nil {
		return
	}
	pending, err := p.api.ListTasksPendingWorktree(ctx, 100)
	if err != nil {
		slog.Warn("list tasks pending worktree", "err", err)
		return
	}
	for _, row := range pending {
		repoID := strings.TrimSpace(row.RepositoryID)
		if repoID == "" {
			continue
		}
		p.provisionOne(ctx, row.TaskID, repoID)
	}
}

func (p *WorktreeProvisioner) provisionOne(ctx context.Context, taskID, repositoryID string) {
	taskID = strings.TrimSpace(taskID)
	repositoryID = strings.TrimSpace(repositoryID)
	if taskID == "" || repositoryID == "" || p.api == nil {
		return
	}
	p.mu.Lock()
	if _, ok := p.inflight[taskID]; ok {
		p.mu.Unlock()
		return
	}
	p.inflight[taskID] = struct{}{}
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		delete(p.inflight, taskID)
		p.mu.Unlock()
	}()

	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "composition.WorktreeProvisioner.provisionOne",
		"task_id", taskID, "repository_id", repositoryID)

	cur, err := p.api.Get(ctx, taskID)
	if err != nil {
		slog.Warn("worktree provision get task", "task_id", taskID, "err", err)
		return
	}
	if cur == nil {
		return
	}
	if cur.WorktreeID != nil && strings.TrimSpace(*cur.WorktreeID) != "" {
		return
	}
	if cur.Status == taskcoredomain.StatusFailed || cur.Status == taskcoredomain.StatusDone {
		return
	}

	wt, err := p.api.AllocateTaskWorktree(ctx, repositoryID, taskID)
	if err != nil {
		slog.Warn("worktree allocate failed", "task_id", taskID, "repository_id", repositoryID, "err", err)
		p.failTask(ctx, taskID, err)
		return
	}
	wtID := wt.ID
	updated, _, err := p.api.taskcore.Update(ctx, taskID, taskcorestore.UpdateTaskInput{
		WorktreeID: &wtID,
	}, taskcoredomain.ActorAgent)
	if err != nil {
		slog.Warn("worktree bind update failed", "task_id", taskID, "err", err)
		p.failTask(ctx, taskID, err)
		return
	}
	if updated == nil {
		return
	}
	_ = realtime.PublishEnrichedTaskUpdated(ctx, p.hub, func(ctx context.Context, id string) (any, error) {
		return p.api.Get(ctx, id)
	}, taskID)
	now := time.Now().UTC()
	// prevStatus empty → treat bind as becoming queue-eligible (create deferred notify).
	p.api.applyNotifyDecision(ctx, *updated, scheduling.DecideNotifyAfterReadyTransition("", updated, false, now))
}

func (p *WorktreeProvisioner) failTask(ctx context.Context, taskID string, cause error) {
	failed := taskcoredomain.StatusFailed
	updated, _, err := p.api.taskcore.Update(ctx, taskID, taskcorestore.UpdateTaskInput{
		Status: &failed,
	}, taskcoredomain.ActorAgent)
	if err != nil {
		slog.Warn("mark task failed after worktree provision error",
			"task_id", taskID, "err", err, "cause", cause)
		return
	}
	_ = realtime.PublishEnrichedTaskUpdated(ctx, p.hub, func(ctx context.Context, id string) (any, error) {
		return p.api.Get(ctx, id)
	}, taskID)
	if updated != nil {
		p.api.applyNotifyDecision(ctx, *updated, scheduling.DecideNotifyAfterReadyTransition(
			taskcoredomain.StatusReady, updated, false, time.Now().UTC()))
	}
}
