package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	composehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/handler"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
)

const instantiateQueueSize = 32

type instantiateWorker struct {
	store   HandlerStore
	prepare func(ctx context.Context, payload json.RawMessage) (*taskcorehandler.PreparedComposeCreate, error)
	create  func(ctx context.Context, prepared *taskcorehandler.PreparedComposeCreate, number *int, by taskcoredomain.Actor) (*taskcoredomain.Task, error)
	jobs    chan composehandler.InstantiateJob

	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}
}

type taskNumberAllocator interface {
	AllocateNextTaskNumbers(ctx context.Context, projectID string, k int) ([]int, error)
}

//funclogmeasure:skip category=hot-path reason="Handler wiring; Run emits operation traces."
func (h *Handler) ensureInstantiateWorker() composehandler.EnqueueInstantiateFunc {
	if h.enqueueInstantiate != nil {
		return h.enqueueInstantiate
	}
	tc := h.taskcoreHandler()
	w := &instantiateWorker{
		store: h.store,
		prepare: func(ctx context.Context, payload json.RawMessage) (*taskcorehandler.PreparedComposeCreate, error) {
			compose, err := taskcorehandler.DecodeComposePayload(payload)
			if err != nil {
				return nil, err
			}
			return tc.PrepareComposeCreate(ctx, compose, taskcorehandler.CreateTaskComposeOpts{
				StripDependsOn:          true,
				OmitPastPickupNotBefore: true,
				InstantiateFromTemplate: true,
			})
		},
		create: func(ctx context.Context, prepared *taskcorehandler.PreparedComposeCreate, number *int, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
			return tc.CreateFromPrepared(ctx, prepared, number, by)
		},
		jobs:   make(chan composehandler.InstantiateJob, instantiateQueueSize),
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
	h.instantiateWorker = w
	h.enqueueInstantiate = w.Enqueue
	go w.Run(context.Background())
	return h.enqueueInstantiate
}

func (w *instantiateWorker) Enqueue(job composehandler.InstantiateJob) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.instantiateWorker.Enqueue",
		"items", len(job.Items))
	if w == nil || len(job.Items) == 0 {
		return false
	}
	select {
	case <-w.stopCh:
		return false
	default:
	}
	select {
	case w.jobs <- job:
		return true
	default:
		slog.Warn("instantiate queue full; rejecting job", "items", len(job.Items))
		return false
	}
}

func (w *instantiateWorker) Stop() {
	if w == nil {
		return
	}
	w.stopOnce.Do(func() { close(w.stopCh) })
	select {
	case <-w.doneCh:
	case <-time.After(15 * time.Second):
		slog.Warn("instantiate worker stop timed out")
	}
}

func (w *instantiateWorker) Run(ctx context.Context) {
	defer close(w.doneCh)
	slog.Info("instantiate worker started", "cmd", calltrace.LogCmd,
		"operation", "handler.instantiateWorker.Run")
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case job := <-w.jobs:
			w.processJob(ctx, job)
		}
	}
}

func (w *instantiateWorker) processJob(ctx context.Context, job composehandler.InstantiateJob) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.instantiateWorker.processJob",
		"items", len(job.Items))
	successCounts := make(map[string]int)
	for _, item := range job.Items {
		prepared, err := w.prepare(ctx, item.Payload)
		if err != nil {
			slog.Warn("instantiate prepare failed", "template_id", item.TemplateID, "err", err)
			continue
		}
		numbers := w.allocateNumbers(ctx, prepared, item.Count)
		for i := 0; i < item.Count; i++ {
			var numPtr *int
			if i < len(numbers) {
				n := numbers[i]
				numPtr = &n
			}
			task, err := w.create(ctx, prepared, numPtr, job.Actor)
			if err != nil {
				slog.Warn("instantiate create failed", "template_id", item.TemplateID, "err", err)
				continue
			}
			if task != nil {
				successCounts[item.TemplateID]++
			}
		}
	}
	if len(successCounts) == 0 || w.store == nil {
		return
	}
	bumpCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := w.store.IncrementTemplateInstantiateCounts(bumpCtx, successCounts); err != nil {
		slog.Warn("increment template instantiate_count after async job", "err", err)
	}
}

func (w *instantiateWorker) allocateNumbers(
	ctx context.Context,
	prepared *taskcorehandler.PreparedComposeCreate,
	count int,
) []int {
	if prepared == nil || count < 1 {
		return nil
	}
	alloc, ok := w.store.(taskNumberAllocator)
	if !ok {
		return nil
	}
	pid := ""
	if prepared.Input.ProjectID != nil {
		pid = strings.TrimSpace(*prepared.Input.ProjectID)
	}
	if pid == "" {
		return nil
	}
	nums, err := alloc.AllocateNextTaskNumbers(ctx, pid, count)
	if err != nil {
		slog.Warn("preallocate task numbers failed; falling back to per-create allocate",
			"project_id", pid, "err", err)
		return nil
	}
	return nums
}
