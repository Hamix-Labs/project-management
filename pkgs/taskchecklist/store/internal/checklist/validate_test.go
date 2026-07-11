package checklist

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
)

func TestSetDoneWithEvidence_rejectsEmptyEvidence(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	ctx := t.Context()

	tskRow := taskmodel.FromDomainTask(taskcoredomain.Task{
		ID: "task-1", Title: "t", InitialPrompt: "p", Status: taskcoredomain.StatusReady, Priority: taskcoredomain.PriorityMedium,
	})
	if err := db.WithContext(ctx).Create(&tskRow).Error; err != nil {
		t.Fatal(err)
	}
	itRow := checklistmodel.FromDomainTaskChecklistItem(checklistdomain.TaskChecklistItem{ID: "item-1", TaskID: tskRow.ID, SortOrder: 1, Text: "criterion"})
	if err := db.WithContext(ctx).Create(&itRow).Error; err != nil {
		t.Fatal(err)
	}
	_, err := SetDoneWithEvidence(ctx, db, tskRow.ID, itRow.ID, "", checklistdomain.VerifierAgentSelf, "", "", taskcoredomain.ActorAgent)
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("got %v, want ErrInvalidInput", err)
	}
}
