package templates

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPatch_templatePayloadRoundTrip(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC()
	if err := model.AutoMigrateAll(db); err != nil {
		t.Fatal(err)
	}

	initial, err := json.Marshal(map[string]any{
		"title":          "Refactor module",
		"priority":       "high",
		"status":         "ready",
		"project_id":     "proj-1",
		"worktree_id":    "wt-1",
		"initial_prompt": "<p>Find a module with poor test coverage.</p>",
		"checklist_items": []map[string]any{
			{"text": "Module named in report"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	row := model.TaskTemplate{
		ID:          "tmpl-1",
		Name:        "Refactor module",
		PayloadJSON: initial,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.WithContext(ctx).Create(&row).Error; err != nil {
		t.Fatal(err)
	}

	patchPayload, err := json.Marshal(map[string]any{
		"title":          "Refactor module",
		"priority":       "high",
		"status":         "ready",
		"project_id":     "proj-1",
		"worktree_id":    "wt-1",
		"initial_prompt": "<p>Find a module with poor test coverage.</p>",
		"checklist_items": []map[string]any{
			{"text": "Module named in report"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	name := row.Name
	detail, err := Patch(ctx, db, row.ID, &name, patchPayload)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(detail.Payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["title"] != "Refactor module" {
		t.Fatalf("title = %v", decoded["title"])
	}
	prompt, _ := decoded["initial_prompt"].(string)
	if !strings.Contains(prompt, "poor test coverage") {
		t.Fatalf("initial_prompt = %q", prompt)
	}
}
