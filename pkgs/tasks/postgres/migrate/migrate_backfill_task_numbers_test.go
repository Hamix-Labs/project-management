package migrate

import (
	"context"
	"testing"
	"time"

	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	taskmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openTaskNumbersDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := autoMigrateStoreModels(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestMigrateBackfillTaskNumbers_assignsDensePerProject(t *testing.T) {
	db := openTaskNumbersDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	repoID := createRepo(t, db, now)
	proj := createProject(t, db, projectsdomain.Project{
		ID: uuid.NewString(), Name: "P", Status: projectsdomain.ProjectStatusActive,
		RepositoryID: &repoID, NextTaskNumber: 1, CreatedAt: now, UpdatedAt: now,
	})

	t1, t2 := uuid.NewString(), uuid.NewString()
	for i, id := range []string{t1, t2} {
		at := now.Add(time.Duration(i) * time.Minute)
		if err := db.Create(&taskmodel.Task{
			ID: id, Title: "t", InitialPrompt: "p", Status: "ready", Priority: "medium",
			ProjectID: &proj.ID, Runner: "cursor",
		}).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(
			`INSERT INTO task_events (task_id, seq, at, type, "by", data_json) VALUES (?, 1, ?, 'task_created', 'user', '{}')`,
			id, at,
		).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := migrateBackfillTaskNumbers(ctx, db); err != nil {
		t.Fatal(err)
	}

	var got1, got2 taskmodel.Task
	if err := db.First(&got1, "id = ?", t1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&got2, "id = ?", t2).Error; err != nil {
		t.Fatal(err)
	}
	if got1.Number == nil || *got1.Number != 1 {
		t.Fatalf("task1 number = %v want 1", got1.Number)
	}
	if got2.Number == nil || *got2.Number != 2 {
		t.Fatalf("task2 number = %v want 2", got2.Number)
	}
	p := loadProject(t, db, proj.ID)
	if p.NextTaskNumber != 3 {
		t.Fatalf("next_task_number = %d want 3", p.NextTaskNumber)
	}

	// Idempotent.
	if err := migrateBackfillTaskNumbers(ctx, db); err != nil {
		t.Fatal(err)
	}
	p2 := loadProject(t, db, proj.ID)
	if p2.NextTaskNumber != 3 {
		t.Fatalf("after re-run next_task_number = %d want 3", p2.NextTaskNumber)
	}
}

func TestMigrateBackfillTaskNumbers_skipsNullProjectTasks(t *testing.T) {
	db := openTaskNumbersDB(t)
	ctx := context.Background()
	id := uuid.NewString()
	if err := db.Create(&taskmodel.Task{
		ID: id, Title: "orphan", InitialPrompt: "p", Status: "ready", Priority: "medium",
		Runner: "cursor",
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrateBackfillTaskNumbers(ctx, db); err != nil {
		t.Fatal(err)
	}
	var got taskmodel.Task
	if err := db.First(&got, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	if got.Number != nil {
		t.Fatalf("null-project task number = %v want nil", got.Number)
	}
}
