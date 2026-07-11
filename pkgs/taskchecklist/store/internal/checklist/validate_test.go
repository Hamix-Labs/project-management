package checklist

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	checklistmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store/model"
	taskmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

func TestSetDoneWithEvidence_rejectsEmptyEvidence(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	ctx := t.Context()

	tskRow := taskmodel.FromDomainTask(domain.Task{
		ID: "task-1", Title: "t", InitialPrompt: "p", Status: domain.StatusReady, Priority: domain.PriorityMedium,
	})
	if err := db.WithContext(ctx).Create(&tskRow).Error; err != nil {
		t.Fatal(err)
	}
	itRow := checklistmodel.FromDomainTaskChecklistItem(checklistdomain.TaskChecklistItem{ID: "item-1", TaskID: tskRow.ID, SortOrder: 1, Text: "criterion"})
	if err := db.WithContext(ctx).Create(&itRow).Error; err != nil {
		t.Fatal(err)
	}
	_, err := SetDoneWithEvidence(ctx, db, tskRow.ID, itRow.ID, "", checklistdomain.VerifierAgentSelf, "", "", domain.ActorAgent)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v, want ErrInvalidInput", err)
	}
}
