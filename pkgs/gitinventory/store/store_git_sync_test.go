package store

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcoremodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestWorktreeStaleMap_usesCycleEndedAt(t *testing.T) {
	s, ctx, _ := gitTestStore(t)
	main := initGitRepo(t)
	created, err := s.CreateGlobalGitRepository(ctx, contract.CreateGitRepositoryInput{Path: main})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}

	wts, err := s.ListGitWorktreesByRepo(ctx, created.ID)
	if err != nil {
		t.Fatalf("ListGitWorktreesByRepo: %v", err)
	}
	got, err := s.WorktreeStaleMap(ctx, wts, time.Now().UTC())
	if err != nil {
		t.Fatalf("WorktreeStaleMap: %v", err)
	}
	if len(wts) == 0 {
		t.Fatal("expected at least the main worktree")
	}
	for _, wt := range wts {
		if _, ok := got[wt.ID]; !ok {
			t.Fatalf("missing stale entry for worktree %s", wt.ID)
		}
		if wt.IsMain && got[wt.ID] {
			t.Fatalf("main worktree %s should not be stale", wt.ID)
		}
	}
}

func TestWorktreeStaleMap_marksIdleWorktreeStale(t *testing.T) {
	s, ctx, _ := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, contract.CreateGitRepositoryInput{Path: main})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-stale")
	wt, err := s.CreateGitWorktreeForRepo(ctx, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-stale",
		CreateBranch: true,
	})
	if err != nil {
		t.Fatalf("CreateGitWorktreeForRepo: %v", err)
	}

	taskID := "task-stale-1"
	wtID := wt.ID
	task := taskcoremodel.Task{
		ID:            taskID,
		Title:         "stale task",
		InitialPrompt: "prompt",
		Status:        taskcoredomain.StatusDone,
		Priority:      taskcoredomain.PriorityMedium,
		Tags:          datatypes.JSONSlice[string]{},
		Runner:        "cursor",
		RunnerConfig:  datatypes.JSON([]byte("{}")),
		WorktreeID:    &wtID,
	}
	if err := s.db.WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}
	ended := time.Now().UTC().Add(-48 * time.Hour)
	cycle := cyclesmodel.TaskCycle{
		ID:          "cycle-stale-1",
		TaskID:      taskID,
		AttemptSeq:  1,
		Status:      cyclesdomain.CycleStatusSucceeded,
		StartedAt:   ended.Add(-time.Hour),
		EndedAt:     &ended,
		TriggeredBy: "user",
		MetaJSON:    datatypes.JSON([]byte("{}")),
	}
	if err := s.db.WithContext(ctx).Create(&cycle).Error; err != nil {
		t.Fatalf("create cycle: %v", err)
	}

	wts, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListGitWorktreesByRepo: %v", err)
	}
	got, err := s.WorktreeStaleMap(ctx, wts, time.Now().UTC())
	if err != nil {
		t.Fatalf("WorktreeStaleMap: %v", err)
	}
	if !got[wt.ID] {
		t.Fatalf("expected worktree %s to be stale", wt.ID)
	}
}

func TestWorktreeStaleMap_activeTaskNotStale(t *testing.T) {
	s, ctx, _ := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, contract.CreateGitRepositoryInput{Path: main})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-active")
	wt, err := s.CreateGitWorktreeForRepo(ctx, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-active",
		CreateBranch: true,
	})
	if err != nil {
		t.Fatalf("CreateGitWorktreeForRepo: %v", err)
	}

	taskID := "task-active-1"
	wtID := wt.ID
	task := taskcoremodel.Task{
		ID:            taskID,
		Title:         "active task",
		InitialPrompt: "prompt",
		Status:        taskcoredomain.StatusRunning,
		Priority:      taskcoredomain.PriorityMedium,
		Tags:          datatypes.JSONSlice[string]{},
		Runner:        "cursor",
		RunnerConfig:  datatypes.JSON([]byte("{}")),
		WorktreeID:    &wtID,
	}
	if err := s.db.WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}

	wts, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListGitWorktreesByRepo: %v", err)
	}
	got, err := s.WorktreeStaleMap(ctx, wts, time.Now().UTC())
	if err != nil {
		t.Fatalf("WorktreeStaleMap: %v", err)
	}
	if got[wt.ID] {
		t.Fatalf("expected worktree %s not to be stale while a task is running", wt.ID)
	}
}

func TestWorktreeStaleMap_queryCountIndependentOfWorktreeCount(t *testing.T) {
	s, ctx, _ := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, contract.CreateGitRepositoryInput{Path: main})
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}

	for i := 0; i < 5; i++ {
		wtPath := filepath.Join(filepath.Dir(main), fmt.Sprintf("wt-q-%d", i))
		_, err := s.CreateGitWorktreeForRepo(ctx, repo.ID, CreateGitWorktreeInput{
			Path:         wtPath,
			Branch:       fmt.Sprintf("feature-q-%d", i),
			CreateBranch: true,
		})
		if err != nil {
			t.Fatalf("CreateGitWorktreeForRepo %d: %v", i, err)
		}
	}

	wts, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListGitWorktreesByRepo: %v", err)
	}
	if len(wts) < 6 {
		t.Fatalf("expected >= 6 worktrees (main + 5), got %d", len(wts))
	}

	const callbackName = "test:worktree_stale_map_query_count"
	var queryCount int
	if err := s.db.Callback().Query().Before("gorm:query").Register(callbackName, func(db *gorm.DB) {
		queryCount++
	}); err != nil {
		t.Fatalf("register callback: %v", err)
	}
	t.Cleanup(func() {
		_ = s.db.Callback().Query().Remove(callbackName)
	})

	queryCount = 0
	if _, err := s.WorktreeStaleMap(ctx, wts, time.Now().UTC()); err != nil {
		t.Fatalf("WorktreeStaleMap: %v", err)
	}
	firstCount := queryCount
	// Set queries only — must not grow with worktree count (old path was 2n).
	if firstCount < 1 || firstCount > 3 {
		t.Fatalf("WorktreeStaleMap ran %d queries, want 1–3 (O(1) set queries)", firstCount)
	}

	wtPath := filepath.Join(filepath.Dir(main), "wt-q-extra")
	if _, err := s.CreateGitWorktreeForRepo(ctx, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-q-extra",
		CreateBranch: true,
	}); err != nil {
		t.Fatalf("CreateGitWorktreeForRepo extra: %v", err)
	}
	wts2, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListGitWorktreesByRepo: %v", err)
	}
	queryCount = 0
	if _, err := s.WorktreeStaleMap(ctx, wts2, time.Now().UTC()); err != nil {
		t.Fatalf("WorktreeStaleMap: %v", err)
	}
	if queryCount != firstCount {
		t.Fatalf("WorktreeStaleMap query count changed with more worktrees: %d → %d", firstCount, queryCount)
	}
}
