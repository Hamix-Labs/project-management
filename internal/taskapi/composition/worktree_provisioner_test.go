package composition_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory"
	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	projectsstore "github.com/AlexsanderHamir/Hamix/pkgs/projects/store"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func TestWorktreeProvisioner_allocatesAndNotifies(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Setenv(gitinventory.EnvManagedWorktreeRoot, t.TempDir())
	api := composition.NewAPI(tasktestdb.OpenSQLite(t))
	q := agents.NewMemoryQueue(8)
	api.SetReadyTaskNotifier(q)

	repo, err := api.CreateGlobalGitRepository(ctx, gitinventorystore.CreateGitRepositoryInput{Path: seedRemoteMainRepo(t)})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	proj, err := api.CreateProject(ctx, projectsstore.CreateProjectInput{
		Name:         "provision-proj",
		RepositoryID: &repo.ID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	hub := realtime.NewSSEHub()
	prov := composition.NewWorktreeProvisioner(api, hub)
	api.SetWorktreeProvisioner(prov)
	go prov.Run(ctx)
	defer prov.Stop()

	created, err := api.Create(ctx, taskcorestore.CreateTaskInput{
		Title:     "async-wt",
		Priority:  taskcoredomain.PriorityMedium,
		Status:    taskcoredomain.StatusReady,
		ProjectID: &proj.ID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.WorktreeID != nil {
		t.Fatalf("create should leave worktree_id nil, got %v", created.WorktreeID)
	}

	deadline := time.Now().Add(10 * time.Second)
	var got *taskcoredomain.Task
	for time.Now().Before(deadline) {
		got, err = api.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.WorktreeID != nil && *got.WorktreeID != "" {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if got == nil || got.WorktreeID == nil || *got.WorktreeID == "" {
		t.Fatal("expected provisioner to bind worktree_id")
	}

	select {
	case task := <-q.Recv():
		q.AckAfterRecv(task.ID)
		if task.ID != created.ID {
			t.Fatalf("notified %q want %q", task.ID, created.ID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("expected NotifyReady after provision")
	}
}

func TestWorktreeProvisioner_allocateFailureMarksFailed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Setenv(gitinventory.EnvManagedWorktreeRoot, t.TempDir())
	api := composition.NewAPI(tasktestdb.OpenSQLite(t))

	mainPath := seedRemoteMainRepo(t)
	repo, err := api.CreateGlobalGitRepository(ctx, gitinventorystore.CreateGitRepositoryInput{Path: mainPath})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	// Break the on-disk checkout after registration so allocate fails closed.
	if err := os.RemoveAll(mainPath); err != nil {
		t.Fatalf("remove main checkout: %v", err)
	}
	proj, err := api.CreateProject(ctx, projectsstore.CreateProjectInput{
		Name:         "fail-proj",
		RepositoryID: &repo.ID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	hub := realtime.NewSSEHub()
	prov := composition.NewWorktreeProvisioner(api, hub)
	api.SetWorktreeProvisioner(prov)
	go prov.Run(ctx)
	defer prov.Stop()

	created, err := api.Create(ctx, taskcorestore.CreateTaskInput{
		Title:     "fail-wt",
		Priority:  taskcoredomain.PriorityMedium,
		Status:    taskcoredomain.StatusReady,
		ProjectID: &proj.ID,
	}, taskcoredomain.ActorUser)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	var got *taskcoredomain.Task
	for time.Now().Before(deadline) {
		got, err = api.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.Status == taskcoredomain.StatusFailed {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if got == nil || got.Status != taskcoredomain.StatusFailed {
		t.Fatalf("status=%v want failed", got)
	}
	if got.WorktreeID != nil {
		t.Fatalf("failed provision should leave worktree_id nil, got %v", got.WorktreeID)
	}
}
