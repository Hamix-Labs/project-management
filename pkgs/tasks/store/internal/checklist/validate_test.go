package checklist

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/model"
)

func TestSetDoneWithEvidence_rejectsEmptyEvidence(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	ctx := t.Context()

	tskRow := model.FromDomainTask(domain.Task{
		ID: "task-1", Title: "t", InitialPrompt: "p", Status: domain.StatusReady, Priority: domain.PriorityMedium,
	})
	if err := db.WithContext(ctx).Create(&tskRow).Error; err != nil {
		t.Fatal(err)
	}
	itRow := model.FromDomainTaskChecklistItem(domain.TaskChecklistItem{ID: "item-1", TaskID: tskRow.ID, SortOrder: 1, Text: "criterion"})
	if err := db.WithContext(ctx).Create(&itRow).Error; err != nil {
		t.Fatal(err)
	}
	_, err := SetDoneWithEvidence(ctx, db, tskRow.ID, itRow.ID, "", domain.VerifierAgentSelf, "", "", domain.ActorAgent)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v, want ErrInvalidInput", err)
	}
}
