package devsim

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	"github.com/google/uuid"
	"log/slog"
	"math/rand/v2"
)

const devsimTaskIDPrefix = "hamix-devsim-"

// RunLifecycleOnce either creates a prefixed dev task or deletes one (no children), then calls publish.
func RunLifecycleOnce(ctx context.Context, st *composition.API, publish func(ChangeKind, string)) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "devsim.RunLifecycleOnce")
	if st == nil || publish == nil {
		return
	}
	if rand.IntN(2) == 0 {
		tryCreateDevsimTask(ctx, st, publish)
		return
	}
	tryDeleteDevsimTask(ctx, st, publish)
}

func tryCreateDevsimTask(ctx context.Context, st *composition.API, publish func(ChangeKind, string)) {
	id := devsimTaskIDPrefix + uuid.NewString()
	t, err := st.Create(ctx, taskcorestore.CreateTaskInput{
		ID:            id,
		Title:         "Dev sim task",
		InitialPrompt: "<p>Synthetic task for UI / SSE exercise.</p>",
		Status:        taskcoredomain.StatusReady,
		Priority:      taskcoredomain.PriorityMedium,
	}, taskcoredomain.ActorAgent)
	if err != nil {
		slog.Debug("sse dev lifecycle create skipped", "cmd", calltrace.LogCmd, "operation", "devsim.lifecycle_create", "err", err)
		return
	}
	publish(ChangeCreated, t.ID)
}

func tryDeleteDevsimTask(ctx context.Context, st *composition.API, publish func(ChangeKind, string)) {
	tasks, err := st.ListDevsimTasks(ctx, devsimTaskIDPrefix+"%")
	if err != nil {
		slog.Debug("sse dev lifecycle list skipped", "cmd", calltrace.LogCmd, "operation", "devsim.lifecycle_list", "err", err)
		return
	}
	if len(tasks) == 0 {
		return
	}
	for _, i := range rand.Perm(len(tasks)) {
		id := tasks[i].ID
		deletedIDs, err := st.Delete(ctx, id, taskcoredomain.ActorAgent)
		if err != nil {
			continue
		}
		for _, deletedID := range deletedIDs {
			publish(ChangeDeleted, deletedID)
		}
		return
	}
}
