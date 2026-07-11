package contract

import (
	"time"

	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// CreateChecklistItemInput is one criterion seeded at task create.
type CreateChecklistItemInput = checklistcontract.CreateChecklistItemInput

// VerifyCommandInput is a verify command on checklist create/update.
type VerifyCommandInput = checklistcontract.VerifyCommandInput

// CreateTaskInput is the task creation payload.
type CreateTaskInput struct {
	ID                    string
	DraftID               string
	Title                 string
	InitialPrompt         string
	Status                domain.Status
	Priority              domain.Priority
	ProjectID             *string
	ProjectContextItemIDs []string
	Runner                string
	CursorModel           string
	PickupNotBefore       *time.Time
	Tags                  []string
	Milestone             *string
	Gate                  *domain.TaskGate
	DependsOn             []domain.DependencyEdge
	ChecklistItems        []CreateChecklistItemInput
	WorktreeID            *string
}

// PickupNotBeforePatch updates pickup_not_before when non-nil.
type PickupNotBeforePatch struct {
	Clear bool
	At    time.Time
}

// UpdateTaskInput is the task patch payload.
type UpdateTaskInput struct {
	Title                 *string
	InitialPrompt         *string
	Status                *domain.Status
	Priority              *domain.Priority
	Project               *ProjectFieldPatch
	ProjectContextItemIDs *[]string
	PickupNotBefore       *PickupNotBeforePatch
	CursorModel           *string
	Tags                  *[]string
	Milestone             *string
	Gate                  **domain.TaskGate
	DependsOn             *[]domain.DependencyEdge
	PendingRetry          *domain.PendingRetry
	ClearPendingRetry     bool
	WorktreeID            *string
}

// ListFilter optionally restricts flat task listing.
type ListFilter struct {
	Tag       *string
	Milestone *string
}

// ProjectFieldPatch updates project_id when non-nil.
type ProjectFieldPatch struct {
	Clear bool
	ID    string
}

// RequestRetryInput is the store payload for operator retry after failure.
type RequestRetryInput struct {
	TaskID        string
	Mode          domain.RetryMode
	ParentCycleID string
}
