// Package namedpayload implements shared CRUD for named JSON-payload
// entities (task drafts and task templates).
package namedpayload

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/store/model"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Summary is the listing-row shape for drafts.
type Summary = contract.DraftSummary

// TemplateSummary is the listing-row shape for task templates.
type TemplateSummary = contract.TemplateSummary

// Detail is the GET-by-id body shape for drafts and templates.
type Detail = contract.DraftDetail

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func clampLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func saveRow(
	ctx context.Context,
	db *gorm.DB,
	id, name string,
	payload json.RawMessage,
	nameRequiredMsg string,
	opSave string,
	logOp string,
	saveErr string,
	updateErr string,
	newRow func(string, string, datatypes.JSON, time.Time) model.TaskDraft,
) (*Summary, error) {
	defer storekernel.DeferLatency(opSave)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", logOp)
	id = storekernel.ResolveID(id)
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: %s", taskcoredomain.ErrInvalidInput, nameRequiredMsg)
	}
	normalized, err := storekernel.NormalizeJSONObject(payload, "payload", taskcoredomain.ErrInvalidInput)
	if err != nil {
		return nil, err
	}
	payload = normalized
	now := time.Now().UTC()
	row := newRow(id, name, datatypes.JSON(payload), now)
	if err := db.WithContext(ctx).Where("id = ?", id).FirstOrCreate(&row).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", saveErr, err)
	}
	if err := db.WithContext(ctx).Model(&model.TaskDraft{}).Where("id = ?", id).Updates(map[string]any{
		"name":         name,
		"payload_json": datatypes.JSON(payload),
		"updated_at":   now,
	}).Error; err != nil {
		return nil, storekernel.MapPayloadPersistenceError(fmt.Errorf("%s: %w", updateErr, err), taskcoredomain.ErrInvalidInput)
	}
	return &Summary{ID: id, Name: name, UpdatedAt: now, CreatedAt: row.CreatedAt}, nil
}

func saveTemplateRow(
	ctx context.Context,
	db *gorm.DB,
	id, name string,
	payload json.RawMessage,
	nameRequiredMsg string,
	opSave string,
	logOp string,
	saveErr string,
	updateErr string,
	newRow func(string, string, datatypes.JSON, time.Time) model.TaskTemplate,
) (*TemplateSummary, error) {
	defer storekernel.DeferLatency(opSave)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", logOp)
	id = storekernel.ResolveID(id)
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: %s", taskcoredomain.ErrInvalidInput, nameRequiredMsg)
	}
	normalized, err := storekernel.NormalizeJSONObject(payload, "payload", taskcoredomain.ErrInvalidInput)
	if err != nil {
		return nil, err
	}
	payload = normalized
	now := time.Now().UTC()
	row := newRow(id, name, datatypes.JSON(payload), now)
	if err := db.WithContext(ctx).Where("id = ?", id).FirstOrCreate(&row).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", saveErr, err)
	}
	if err := db.WithContext(ctx).Model(&model.TaskTemplate{}).Where("id = ?", id).Updates(map[string]any{
		"name":         name,
		"payload_json": datatypes.JSON(payload),
		"updated_at":   now,
	}).Error; err != nil {
		return nil, storekernel.MapPayloadPersistenceError(fmt.Errorf("%s: %w", updateErr, err), taskcoredomain.ErrInvalidInput)
	}
	return &TemplateSummary{
		ID:               id,
		Name:             name,
		UpdatedAt:        now,
		CreatedAt:        row.CreatedAt,
		PrimaryTag:       primaryTagFromPayload(datatypes.JSON(payload)),
		InstantiateCount: row.InstantiateCount,
	}, nil
}

func SaveDraft(ctx context.Context, db *gorm.DB, id, name string, payload json.RawMessage) (*Summary, error) {
	return saveRow(ctx, db, id, name, payload,
		"draft name required",
		storekernel.OpSaveDraft,
		"tasks.store.drafts.Save",
		"save draft", "update draft",
		func(id, name string, p datatypes.JSON, now time.Time) model.TaskDraft {
			return model.TaskDraft{ID: id, Name: name, PayloadJSON: p, CreatedAt: now, UpdatedAt: now}
		},
	)
}

func SaveTemplate(ctx context.Context, db *gorm.DB, id, name string, payload json.RawMessage) (*TemplateSummary, error) {
	return saveTemplateRow(ctx, db, id, name, payload,
		"template name required",
		storekernel.OpSaveTemplate,
		"tasks.store.templates.Save",
		"save template", "update template",
		func(id, name string, p datatypes.JSON, now time.Time) model.TaskTemplate {
			return model.TaskTemplate{ID: id, Name: name, PayloadJSON: p, CreatedAt: now, UpdatedAt: now}
		},
	)
}

func ListDrafts(ctx context.Context, db *gorm.DB, limit int) ([]Summary, error) {
	defer storekernel.DeferLatency(storekernel.OpListDrafts)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.drafts.List")
	limit = clampLimit(limit)
	var rows []model.TaskDraft
	if err := db.WithContext(ctx).Order("updated_at DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list drafts: %w", err)
	}
	return summariesFromDraftRows(rows), nil
}

func ListTemplates(ctx context.Context, db *gorm.DB, limit int, q, sort, order, tag string) ([]TemplateSummary, error) {
	defer storekernel.DeferLatency(storekernel.OpListTemplates)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.templates.List")
	limit = clampLimit(limit)
	orderClause := sort + " " + order
	query := db.WithContext(ctx).Model(&model.TaskTemplate{}).Order(orderClause).Limit(limit)
	q = strings.TrimSpace(q)
	if q != "" {
		like := "%" + escapeLike(strings.ToLower(q)) + "%"
		query = query.Where("LOWER(name) LIKE ?", like)
	}
	tag = strings.TrimSpace(tag)
	if tag != "" {
		if isSQLiteDialect(db) {
			query = query.Where("LOWER(json_extract(payload_json, '$.tags[0]')) = LOWER(?)", tag)
		} else {
			query = query.Where("LOWER(payload_json->'tags'->>0) = LOWER(?)", tag)
		}
	}
	var rows []model.TaskTemplate
	if err := query.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	return summariesFromTemplateRows(rows), nil
}

func GetDraft(ctx context.Context, db *gorm.DB, id string) (*Detail, error) {
	return getDraftByID(ctx, db, id)
}

func GetTemplate(ctx context.Context, db *gorm.DB, id string) (*Detail, error) {
	return getTemplateByID(ctx, db, id)
}

func getDraftByID(ctx context.Context, db *gorm.DB, id string) (*Detail, error) {
	defer storekernel.DeferLatency(storekernel.OpGetDraft)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.drafts.Get")
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	var row model.TaskDraft
	if err := db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, storekernel.MapNotFound(err, taskcoredomain.ErrNotFound)
	}
	return detailFromDraft(row), nil
}

func getTemplateByID(ctx context.Context, db *gorm.DB, id string) (*Detail, error) {
	defer storekernel.DeferLatency(storekernel.OpGetTemplate)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.templates.Get")
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	var row model.TaskTemplate
	if err := db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, storekernel.MapNotFound(err, taskcoredomain.ErrNotFound)
	}
	return detailFromTemplate(row), nil
}

func DeleteDraft(ctx context.Context, db *gorm.DB, id string) error {
	return deleteByID(ctx, db, id, storekernel.OpDeleteDraft, "tasks.store.drafts.Delete", "delete draft", &model.TaskDraft{})
}

func DeleteTemplate(ctx context.Context, db *gorm.DB, id string) error {
	return deleteByID(ctx, db, id, storekernel.OpDeleteTemplate, "tasks.store.templates.Delete", "delete template", &model.TaskTemplate{})
}

func deleteByID(ctx context.Context, db *gorm.DB, id string, op string, logOp, deleteErr string, row any) error {
	defer storekernel.DeferLatency(op)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", logOp)
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%w: id", taskcoredomain.ErrInvalidInput)
	}
	res := db.WithContext(ctx).Where("id = ?", id).Delete(row)
	if res.Error != nil {
		return fmt.Errorf("%s: %w", deleteErr, res.Error)
	}
	if res.RowsAffected == 0 {
		return taskcoredomain.ErrNotFound
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by ListDrafts."
func summariesFromDraftRows(rows []model.TaskDraft) []Summary {
	out := make([]Summary, 0, len(rows))
	for _, r := range rows {
		out = append(out, Summary{ID: r.ID, Name: r.Name, UpdatedAt: r.UpdatedAt, CreatedAt: r.CreatedAt})
	}
	return out
}

func IncrementTemplateInstantiateCounts(ctx context.Context, db *gorm.DB, counts map[string]int) error {
	defer storekernel.DeferLatency(storekernel.OpIncrementTemplateInstantiateCounts)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.templates.IncrementInstantiateCounts")
	for id, delta := range counts {
		if delta <= 0 {
			continue
		}
		res := db.WithContext(ctx).Model(&model.TaskTemplate{}).
			Where("id = ?", id).
			UpdateColumn("instantiate_count", gorm.Expr("instantiate_count + ?", delta))
		if res.Error != nil {
			return fmt.Errorf("increment template instantiate_count: %w", res.Error)
		}
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by ListTemplates."
func summariesFromTemplateRows(rows []model.TaskTemplate) []TemplateSummary {
	out := make([]TemplateSummary, 0, len(rows))
	for _, r := range rows {
		out = append(out, templateSummaryFromRow(r))
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by ListTemplates."
func templateSummaryFromRow(r model.TaskTemplate) TemplateSummary {
	s := TemplateSummary{
		ID:               r.ID,
		Name:             r.Name,
		UpdatedAt:        r.UpdatedAt,
		CreatedAt:        r.CreatedAt,
		InstantiateCount: r.InstantiateCount,
	}
	if tag := primaryTagFromPayload(r.PayloadJSON); tag != "" {
		s.PrimaryTag = tag
	}
	return s
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by ListTemplates."
func primaryTagFromPayload(payload datatypes.JSON) string {
	var p struct {
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || len(p.Tags) == 0 {
		return ""
	}
	return strings.TrimSpace(p.Tags[0])
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by ListTemplates."
func isSQLiteDialect(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	return strings.Contains(strings.ToLower(db.Dialector.Name()), "sqlite")
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func detailFromDraft(row model.TaskDraft) *Detail {
	return &Detail{
		ID: row.ID, Name: row.Name, Payload: json.RawMessage(row.PayloadJSON),
		UpdatedAt: row.UpdatedAt, CreatedAt: row.CreatedAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func detailFromTemplate(row model.TaskTemplate) *Detail {
	return &Detail{
		ID: row.ID, Name: row.Name, Payload: json.RawMessage(row.PayloadJSON),
		UpdatedAt: row.UpdatedAt, CreatedAt: row.CreatedAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by ListTemplates."
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
