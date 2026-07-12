package handler

import (
	"context"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"strconv"
	"testing"
)

const maxTaskEventSeqParamBytes = 32

// appendApprovalRequestedEvent appends approval_requested and returns its seq.
// Use after POST /tasks with checklist_items so event numbering stays stable.
func appendApprovalRequestedEvent(t *testing.T, st *composition.API, ctx context.Context, taskID string) int64 {
	t.Helper()
	if err := st.AppendTaskEvent(ctx, taskID, taskeventsdomain.EventApprovalRequested, taskcoredomain.ActorAgent, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	seq, err := st.LastEventSeq(ctx, taskID)
	if err != nil {
		t.Fatal(err)
	}
	return seq
}

func formatEventSeq(seq int64) string {
	return strconv.FormatInt(seq, 10)
}
