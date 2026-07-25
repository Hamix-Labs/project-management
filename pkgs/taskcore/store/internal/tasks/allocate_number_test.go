package tasks

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestAllocateNextTaskNumber_monotonic(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	ctx := context.Background()
	pid := uuid.NewString()
	if err := db.Create(&projectmodel.Project{
		ID: pid, Name: "P", Status: projectsdomain.ProjectStatusActive,
		NextTaskNumber: 1,
	}).Error; err != nil {
		t.Fatal(err)
	}

	for want := 1; want <= 3; want++ {
		var got int
		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			n, err := allocateNextTaskNumber(tx, pid)
			if err != nil {
				return err
			}
			got = n
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("allocate #%d = %d", want, got)
		}
	}
	var p projectmodel.Project
	if err := db.First(&p, "id = ?", pid).Error; err != nil {
		t.Fatal(err)
	}
	if p.NextTaskNumber != 4 {
		t.Fatalf("next = %d want 4", p.NextTaskNumber)
	}
}

func TestCreate_assignsNumberWhenProjectSet(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	ctx := context.Background()
	pid := uuid.NewString()
	if err := db.Create(&projectmodel.Project{
		ID: pid, Name: "P", Status: projectsdomain.ProjectStatusActive,
		NextTaskNumber: 1,
	}).Error; err != nil {
		t.Fatal(err)
	}

	t1, err := Create(ctx, db, CreateInput{
		Title: "one", Priority: domain.PriorityMedium, ProjectID: &pid,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	t2, err := Create(ctx, db, CreateInput{
		Title: "two", Priority: domain.PriorityMedium, ProjectID: &pid,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if t1.Number == nil || *t1.Number != 1 {
		t.Fatalf("t1.Number = %v want 1", t1.Number)
	}
	if t2.Number == nil || *t2.Number != 2 {
		t.Fatalf("t2.Number = %v want 2", t2.Number)
	}
}

func TestCreate_noNumberWithoutProject(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	ctx := context.Background()
	task, err := Create(ctx, db, CreateInput{
		Title: "solo", Priority: domain.PriorityMedium,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	if task.Number != nil {
		t.Fatalf("Number = %v want nil", task.Number)
	}
}

func TestApplyProjectPatch_rejectsMoveWhenNumbered(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	ctx := context.Background()
	p1, p2 := uuid.NewString(), uuid.NewString()
	for _, pid := range []string{p1, p2} {
		if err := db.Create(&projectmodel.Project{
			ID: pid, Name: "P", Status: projectsdomain.ProjectStatusActive,
			NextTaskNumber: 1,
		}).Error; err != nil {
			t.Fatal(err)
		}
	}
	task, err := Create(ctx, db, CreateInput{
		Title: "n", Priority: domain.PriorityMedium, ProjectID: &p1,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = Update(ctx, db, task.ID, UpdateInput{
		Project: &ProjectFieldPatch{ID: p2},
	}, domain.ActorUser)
	if err == nil || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("move numbered task err = %v want ErrInvalidInput", err)
	}
	if err != nil && !strings.Contains(err.Error(), "cannot move numbered task") {
		t.Fatalf("err = %v want cannot move numbered task", err)
	}
}

func TestApplyProjectPatch_rejectsClearWhenNumbered(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	ctx := context.Background()
	pid := uuid.NewString()
	if err := db.Create(&projectmodel.Project{
		ID: pid, Name: "P", Status: projectsdomain.ProjectStatusActive,
		NextTaskNumber: 1,
	}).Error; err != nil {
		t.Fatal(err)
	}
	task, err := Create(ctx, db, CreateInput{
		Title: "n", Priority: domain.PriorityMedium, ProjectID: &pid,
	}, domain.ActorUser)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = Update(ctx, db, task.ID, UpdateInput{
		Project: &ProjectFieldPatch{Clear: true},
	}, domain.ActorUser)
	if err == nil || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("clear numbered task err = %v want ErrInvalidInput", err)
	}
}
