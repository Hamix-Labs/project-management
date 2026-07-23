package store_test

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

func TestCountTasksByWorktreeID(t *testing.T) {
	t.Parallel()
	st := taskcorestore.NewStore(tasktestdb.OpenSQLite(t))
	ctx := context.Background()

	n, err := st.CountTasksByWorktreeID(ctx, "")
	if err != nil || n != 0 {
		t.Fatalf("empty worktree_id: n=%d err=%v", n, err)
	}
	n, err = st.CountTasksByWorktreeID(ctx, "missing-wt")
	if err != nil || n != 0 {
		t.Fatalf("missing: n=%d err=%v", n, err)
	}

	wtID := "wt-count-1"
	for i := 0; i < 2; i++ {
		if _, err := st.Create(ctx, taskcorestore.CreateTaskInput{
			Title:      "count-wt",
			Priority:   domain.PriorityMedium,
			WorktreeID: &wtID,
		}, domain.ActorUser); err != nil {
			t.Fatal(err)
		}
	}
	n, err = st.CountTasksByWorktreeID(ctx, wtID)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("count=%d want 2", n)
	}
}
