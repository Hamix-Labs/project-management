// package internal owns persistence for first-class projects, curated project
// context items, and immutable task context snapshots.
package internal

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	projectmodel "github.com/AlexsanderHamir/Hamix/pkgs/projects/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
	taskdomain "github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/kernel"
	taskmodel "github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateProjectInput is the store input for creating a project.
type CreateProjectInput = contract.CreateProjectInput

// UpdateProjectInput is a partial patch for project metadata.
type UpdateProjectInput = contract.UpdateProjectInput

// CreateContextInput is the store input for appending a project context item.
type CreateContextInput = contract.CreateProjectContextInput

// UpdateContextInput is a partial patch for one project context item.
type UpdateContextInput = contract.UpdateProjectContextInput

// CreateSnapshotInput records the exact context bundle handed to one cycle.
type CreateSnapshotInput struct {
	ID              string
	TaskID          string
	CycleID         string
	ProjectID       string
	ContextJSON     json.RawMessage
	RenderedContext string
	TokenEstimate   int
}

// CreateProject inserts a new active project.
func CreateProject(ctx context.Context, db *gorm.DB, input CreateProjectInput) (domain.Project, error) {
	defer kernel.DeferLatency(kernel.OpCreateProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.CreateProject")
	id := kernel.ResolveID(input.ID)
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.Project{}, fmt.Errorf("%w: project name required", domain.ErrInvalidInput)
	}
	repoID := trimOptional(input.RepositoryID)
	if repoID == nil {
		return domain.Project{}, fmt.Errorf("%w: repository_id required", domain.ErrInvalidInput)
	}
	var repo taskmodel.GitRepository
	if err := db.WithContext(ctx).First(&repo, "id = ?", *repoID).Error; err != nil {
		return domain.Project{}, kernel.MapNotFound(err)
	}
	now := time.Now().UTC()
	drow := domain.Project{
		ID:             id,
		Name:           name,
		Description:    strings.TrimSpace(input.Description),
		Status:         domain.ProjectStatusActive,
		ContextSummary: strings.TrimSpace(input.ContextSummary),
		RepositoryID:   repoID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	row := projectmodel.FromDomainProject(drow)
	if err := db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Project{}, kernel.MapWriteError(err, "duplicate project row")
	}
	return drow, nil
}

// ListProjects returns projects ordered by most recently updated first.
func ListProjects(ctx context.Context, db *gorm.DB, includeArchived bool, limit int) ([]domain.Project, error) {
	defer kernel.DeferLatency(kernel.OpListProjects)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.ListProjects")
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	q := db.WithContext(ctx).Order("updated_at DESC").Limit(limit)
	if !includeArchived {
		q = q.Where("status = ?", domain.ProjectStatusActive)
	}
	var rows []projectmodel.Project
	if err := q.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return projectmodel.ToDomainProjects(rows), nil
}

// GetProject returns one project by id.
func GetProject(ctx context.Context, db *gorm.DB, id string) (domain.Project, error) {
	defer kernel.DeferLatency(kernel.OpGetProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.GetProject")
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.Project{}, fmt.Errorf("%w: project id required", domain.ErrInvalidInput)
	}
	var row projectmodel.Project
	if err := db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		return domain.Project{}, kernel.MapNotFound(err)
	}
	return projectmodel.ToDomainProject(row), nil
}

// UpdateProject applies a partial metadata patch and returns the updated row.
func UpdateProject(ctx context.Context, db *gorm.DB, id string, input UpdateProjectInput) (domain.Project, error) {
	defer kernel.DeferLatency(kernel.OpUpdateProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.UpdateProject")
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.Project{}, fmt.Errorf("%w: project id required", domain.ErrInvalidInput)
	}
	if err := validateProjectPatch(input); err != nil {
		return domain.Project{}, err
	}
	var out domain.Project
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row projectmodel.Project
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", id).Error; err != nil {
			return kernel.MapNotFound(err)
		}
		drow := projectmodel.ToDomainProject(row)
		if err := validateDefaultProjectPatch(drow, input); err != nil {
			return err
		}
		applyProjectPatch(&drow, input)
		drow.UpdatedAt = time.Now().UTC()
		row = projectmodel.FromDomainProject(drow)
		if err := tx.Save(&row).Error; err != nil {
			return kernel.MapWriteError(err, "duplicate project row")
		}
		out = drow
		return nil
	})
	if err != nil {
		return domain.Project{}, err
	}
	return out, nil
}

// DeleteProject removes a project. Tasks referencing it keep running; project_id
// falls back to NULL via ON DELETE SET NULL on tasks.project_id.
func DeleteProject(ctx context.Context, db *gorm.DB, id string) error {
	defer kernel.DeferLatency(kernel.OpDeleteProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.DeleteProject")
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%w: project id required", domain.ErrInvalidInput)
	}
	var row projectmodel.Project
	if err := db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		return kernel.MapNotFound(err)
	}
	if row.IsDefault {
		return fmt.Errorf("%w: default project cannot be deleted", domain.ErrConflict)
	}
	res := db.WithContext(ctx).Delete(&projectmodel.Project{}, "id = ?", id)
	if res.Error != nil {
		return kernel.MapWriteError(res.Error, "duplicate project row")
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CreateContext inserts one context item for a project.
func CreateContext(ctx context.Context, db *gorm.DB, projectID string, input CreateContextInput) (domain.ProjectContextItem, error) {
	defer kernel.DeferLatency(kernel.OpCreateProjectContext)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.CreateContext")
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return domain.ProjectContextItem{}, fmt.Errorf("%w: project id required", domain.ErrInvalidInput)
	}
	id := kernel.ResolveID(input.ID)
	kind := domain.ProjectContextKind(strings.TrimSpace(string(input.Kind)))
	if kind == "" {
		kind = domain.ProjectContextKindNote
	}
	if err := validateContextKind(kind); err != nil {
		return domain.ProjectContextItem{}, err
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return domain.ProjectContextItem{}, fmt.Errorf("%w: context title required", domain.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Body) == "" {
		return domain.ProjectContextItem{}, fmt.Errorf("%w: context body required", domain.ErrInvalidInput)
	}
	actor := input.CreatedBy
	if actor == "" {
		actor = domain.ActorUser
	}
	if err := kernel.ValidateActor(taskdomain.Actor(actor)); err != nil {
		return domain.ProjectContextItem{}, err
	}
	now := time.Now().UTC()
	drow := domain.ProjectContextItem{
		ID:            id,
		ProjectID:     projectID,
		Kind:          kind,
		Title:         title,
		Body:          strings.TrimSpace(input.Body),
		SourceTaskID:  trimOptional(input.SourceTaskID),
		SourceCycleID: trimOptional(input.SourceCycleID),
		CreatedBy:     actor,
		Pinned:        input.Pinned,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	row := projectmodel.FromDomainProjectContextItem(drow)
	if err := db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.ProjectContextItem{}, kernel.MapWriteError(err, "duplicate project row")
	}
	return drow, nil
}

// ListContext returns context items for a project, pinned items first.
func ListContext(ctx context.Context, db *gorm.DB, projectID string, includeUnpinned bool, limit int) ([]domain.ProjectContextItem, error) {
	defer kernel.DeferLatency(kernel.OpListProjectContext)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.ListContext")
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("%w: project id required", domain.ErrInvalidInput)
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	q := db.WithContext(ctx).Where("project_id = ?", projectID).Order("pinned DESC").Order("updated_at DESC").Limit(limit)
	if !includeUnpinned {
		q = q.Where("pinned = ?", true)
	}
	var rows []projectmodel.ProjectContextItem
	if err := q.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list project context: %w", err)
	}
	return projectmodel.ToDomainProjectContextItems(rows), nil
}

// ListContextByIDs returns selected context items for one project in caller order.
func ListContextByIDs(ctx context.Context, db *gorm.DB, projectID string, ids []string) ([]domain.ProjectContextItem, error) {
	defer kernel.DeferLatency(kernel.OpListProjectContext)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.ListContextByIDs")
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("%w: project id required", domain.ErrInvalidInput)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []projectmodel.ProjectContextItem
	if err := db.WithContext(ctx).Where("project_id = ? AND id IN ?", projectID, ids).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list selected project context: %w", err)
	}
	byID := make(map[string]domain.ProjectContextItem, len(rows))
	for _, row := range rows {
		d := projectmodel.ToDomainProjectContextItem(row)
		byID[d.ID] = d
	}
	out := make([]domain.ProjectContextItem, 0, len(ids))
	for _, id := range ids {
		row, ok := byID[strings.TrimSpace(id)]
		if !ok {
			return nil, domain.ErrNotFound
		}
		out = append(out, row)
	}
	return out, nil
}

// UpdateContext applies a partial patch to one context item.
func UpdateContext(ctx context.Context, db *gorm.DB, projectID, itemID string, input UpdateContextInput) (domain.ProjectContextItem, error) {
	defer kernel.DeferLatency(kernel.OpUpdateProjectContext)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.UpdateContext")
	projectID = strings.TrimSpace(projectID)
	itemID = strings.TrimSpace(itemID)
	if projectID == "" || itemID == "" {
		return domain.ProjectContextItem{}, fmt.Errorf("%w: project id and context id required", domain.ErrInvalidInput)
	}
	if err := validateContextPatch(input); err != nil {
		return domain.ProjectContextItem{}, err
	}
	var out domain.ProjectContextItem
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row projectmodel.ProjectContextItem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ? AND project_id = ?", itemID, projectID).Error; err != nil {
			return kernel.MapNotFound(err)
		}
		drow := projectmodel.ToDomainProjectContextItem(row)
		applyContextPatch(&drow, input)
		drow.UpdatedAt = time.Now().UTC()
		row = projectmodel.FromDomainProjectContextItem(drow)
		if err := tx.Save(&row).Error; err != nil {
			return kernel.MapWriteError(err, "duplicate project row")
		}
		out = drow
		return nil
	})
	if err != nil {
		return domain.ProjectContextItem{}, err
	}
	return out, nil
}

// DeleteContext removes one context item.
func DeleteContext(ctx context.Context, db *gorm.DB, projectID, itemID string) error {
	defer kernel.DeferLatency(kernel.OpDeleteProjectContext)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.DeleteContext")
	projectID = strings.TrimSpace(projectID)
	itemID = strings.TrimSpace(itemID)
	if projectID == "" || itemID == "" {
		return fmt.Errorf("%w: project id and context id required", domain.ErrInvalidInput)
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ? AND (source_context_id = ? OR target_context_id = ?)", projectID, itemID, itemID).Delete(&projectmodel.ProjectContextEdge{}).Error; err != nil {
			return kernel.MapWriteError(err, "duplicate project row")
		}
		res := tx.Where("id = ? AND project_id = ?", itemID, projectID).Delete(&projectmodel.ProjectContextItem{})
		if res.Error != nil {
			return kernel.MapWriteError(res.Error, "duplicate project row")
		}
		if res.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		return nil
	})
}

// CreateSnapshot inserts an immutable task context snapshot.
func CreateSnapshot(ctx context.Context, db *gorm.DB, input CreateSnapshotInput) (taskdomain.TaskContextSnapshot, error) {
	defer kernel.DeferLatency(kernel.OpCreateContextSnapshot)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.CreateSnapshot")
	id := kernel.ResolveID(input.ID)
	if strings.TrimSpace(input.TaskID) == "" || strings.TrimSpace(input.CycleID) == "" || strings.TrimSpace(input.ProjectID) == "" {
		return taskdomain.TaskContextSnapshot{}, fmt.Errorf("%w: task_id, cycle_id, and project_id required", domain.ErrInvalidInput)
	}
	if input.TokenEstimate < 0 {
		return taskdomain.TaskContextSnapshot{}, fmt.Errorf("%w: token_estimate must be >= 0", domain.ErrInvalidInput)
	}
	contextJSON, err := kernel.NormalizeJSONObject(input.ContextJSON, "context_json")
	if err != nil {
		return taskdomain.TaskContextSnapshot{}, err
	}
	drow := taskdomain.TaskContextSnapshot{
		ID:              id,
		TaskID:          strings.TrimSpace(input.TaskID),
		CycleID:         strings.TrimSpace(input.CycleID),
		ProjectID:       strings.TrimSpace(input.ProjectID),
		ContextJSON:     json.RawMessage(contextJSON),
		RenderedContext: strings.TrimSpace(input.RenderedContext),
		TokenEstimate:   input.TokenEstimate,
		CreatedAt:       time.Now().UTC(),
	}
	row := taskmodel.FromDomainTaskContextSnapshot(drow)
	if err := db.WithContext(ctx).Create(&row).Error; err != nil {
		return taskdomain.TaskContextSnapshot{}, kernel.MapWriteError(err, "duplicate project row")
	}
	return drow, nil
}

// GetSnapshotForCycle returns the context snapshot recorded for a cycle.
func GetSnapshotForCycle(ctx context.Context, db *gorm.DB, cycleID string) (taskdomain.TaskContextSnapshot, error) {
	defer kernel.DeferLatency(kernel.OpGetContextSnapshot)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.GetSnapshotForCycle")
	cycleID = strings.TrimSpace(cycleID)
	if cycleID == "" {
		return taskdomain.TaskContextSnapshot{}, fmt.Errorf("%w: cycle id required", domain.ErrInvalidInput)
	}
	var row taskmodel.TaskContextSnapshot
	if err := db.WithContext(ctx).First(&row, "cycle_id = ?", cycleID).Error; err != nil {
		return taskdomain.TaskContextSnapshot{}, kernel.MapNotFound(err)
	}
	return taskmodel.ToDomainTaskContextSnapshot(row), nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func validateProjectPatch(input UpdateProjectInput) error {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return fmt.Errorf("%w: project name required", domain.ErrInvalidInput)
	}
	if input.Status != nil {
		switch *input.Status {
		case domain.ProjectStatusActive, domain.ProjectStatusArchived:
		default:
			return fmt.Errorf("%w: invalid project status %q", domain.ErrInvalidInput, *input.Status)
		}
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func validateDefaultProjectPatch(row domain.Project, input UpdateProjectInput) error {
	if !row.IsDefault {
		return nil
	}
	if input.Name != nil && strings.TrimSpace(*input.Name) != domain.DefaultProjectName {
		return fmt.Errorf("%w: default project name cannot be changed", domain.ErrConflict)
	}
	if input.Status != nil && *input.Status != domain.ProjectStatusActive {
		return fmt.Errorf("%w: default project cannot be archived", domain.ErrConflict)
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func applyProjectPatch(row *domain.Project, input UpdateProjectInput) {
	if input.Name != nil {
		row.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		row.Description = strings.TrimSpace(*input.Description)
	}
	if input.Status != nil {
		row.Status = *input.Status
	}
	if input.ContextSummary != nil {
		row.ContextSummary = strings.TrimSpace(*input.ContextSummary)
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func validateContextPatch(input UpdateContextInput) error {
	if input.Kind != nil {
		if err := validateContextKind(*input.Kind); err != nil {
			return err
		}
	}
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return fmt.Errorf("%w: context title required", domain.ErrInvalidInput)
	}
	if input.Body != nil && strings.TrimSpace(*input.Body) == "" {
		return fmt.Errorf("%w: context body required", domain.ErrInvalidInput)
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func applyContextPatch(row *domain.ProjectContextItem, input UpdateContextInput) {
	if input.Kind != nil {
		row.Kind = domain.ProjectContextKind(strings.TrimSpace(string(*input.Kind)))
	}
	if input.Title != nil {
		row.Title = strings.TrimSpace(*input.Title)
	}
	if input.Body != nil {
		row.Body = strings.TrimSpace(*input.Body)
	}
	if input.Pinned != nil {
		row.Pinned = *input.Pinned
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func validateContextKind(kind domain.ProjectContextKind) error {
	trimmed := strings.TrimSpace(string(kind))
	if trimmed == "" {
		return fmt.Errorf("%w: context kind required", domain.ErrInvalidInput)
	}
	if len(trimmed) > 24 {
		return fmt.Errorf("%w: context kind must be 24 characters or fewer", domain.ErrInvalidInput)
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// CreateDefaultProjectForRepo inserts the non-deletable default for a newly registered repo.
// Idempotent: returns the existing default when one is already present.
func CreateDefaultProjectForRepo(ctx context.Context, tx *gorm.DB, repoID string, now time.Time) (domain.Project, error) {
	defer kernel.DeferLatency(kernel.OpCreateProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.CreateDefaultProjectForRepo")
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		return domain.Project{}, fmt.Errorf("%w: repository_id required", domain.ErrInvalidInput)
	}
	var existing projectmodel.Project
	err := tx.WithContext(ctx).
		Where("repository_id = ? AND is_default = ?", repoID, true).
		First(&existing).Error
	if err == nil {
		return projectmodel.ToDomainProject(existing), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Project{}, fmt.Errorf("lookup default project: %w", err)
	}
	now = now.UTC()
	drow := domain.Project{
		ID:             kernel.ResolveID(""),
		Name:           domain.DefaultProjectName,
		Description:    "Built-in project for tasks tied to this repository.",
		Status:         domain.ProjectStatusActive,
		ContextSummary: "Default project for this repository.",
		RepositoryID:   &repoID,
		IsDefault:      true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	row := projectmodel.FromDomainProject(drow)
	if err := tx.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Project{}, kernel.MapWriteError(err, "duplicate default project")
	}
	return drow, nil
}

// GetDefaultProjectForRepository returns the system default project for a repo.
func GetDefaultProjectForRepository(ctx context.Context, db *gorm.DB, repoID string) (domain.Project, error) {
	defer kernel.DeferLatency(kernel.OpGetProject)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.projects.GetDefaultProjectForRepository")
	repoID = strings.TrimSpace(repoID)
	if repoID == "" {
		return domain.Project{}, fmt.Errorf("%w: repository_id required", domain.ErrInvalidInput)
	}
	var row projectmodel.Project
	err := db.WithContext(ctx).
		Where("repository_id = ? AND is_default = ?", repoID, true).
		First(&row).Error
	if err != nil {
		return domain.Project{}, kernel.MapNotFound(err)
	}
	return projectmodel.ToDomainProject(row), nil
}
