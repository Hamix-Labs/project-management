// Package templates owns task_templates persistence for POST/GET/PATCH/DELETE
// /task-templates. Payload writes use storekernel.NormalizeJSONObject like drafts.
package templates

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/store/internal/namedpayload"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Summary = namedpayload.TemplateSummary
type Detail = namedpayload.Detail

//funclogmeasure:skip category=hot-path reason="Thin delegate to namedpayload; operation trace is emitted there."
func Save(ctx context.Context, db *gorm.DB, id, name string, payload json.RawMessage) (*Summary, error) {
	return namedpayload.SaveTemplate(ctx, db, id, name, payload)
}

//funclogmeasure:skip category=hot-path reason="Thin delegate to namedpayload; operation trace is emitted there."
func List(ctx context.Context, db *gorm.DB, limit int, q, sort, order, tag string) ([]Summary, error) {
	return namedpayload.ListTemplates(ctx, db, limit, q, sort, order, tag)
}

//funclogmeasure:skip category=hot-path reason="Thin delegate to namedpayload; operation trace is emitted there."
func Get(ctx context.Context, db *gorm.DB, id string) (*Detail, error) {
	return namedpayload.GetTemplate(ctx, db, id)
}

func Patch(ctx context.Context, db *gorm.DB, id string, name *string, payload json.RawMessage) (*Detail, error) {
	defer storekernel.DeferLatency(storekernel.OpPatchTemplate)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.templates.Patch")
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	if name == nil && payload == nil {
		return nil, fmt.Errorf("%w: no fields to update", taskcoredomain.ErrInvalidInput)
	}
	var row model.TaskTemplate
	if err := db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, storekernel.MapNotFound(err)
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return nil, fmt.Errorf("%w: template name required", taskcoredomain.ErrInvalidInput)
		}
		updates["name"] = trimmed
	}
	if payload != nil {
		normalized, err := storekernel.NormalizeJSONObject(payload, "payload")
		if err != nil {
			return nil, err
		}
		updates["payload_json"] = datatypes.JSON(normalized)
	}
	if err := db.WithContext(ctx).Model(&model.TaskTemplate{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, storekernel.MapPayloadPersistenceError(fmt.Errorf("patch template: %w", err))
	}
	return Get(ctx, db, id)
}

//funclogmeasure:skip category=hot-path reason="Thin delegate to namedpayload; operation trace is emitted there."
func Delete(ctx context.Context, db *gorm.DB, id string) error {
	return namedpayload.DeleteTemplate(ctx, db, id)
}

//funclogmeasure:skip category=hot-path reason="Thin delegate to namedpayload; operation trace is emitted there."
func IncrementInstantiateCounts(ctx context.Context, db *gorm.DB, counts map[string]int) error {
	return namedpayload.IncrementTemplateInstantiateCounts(ctx, db, counts)
}
