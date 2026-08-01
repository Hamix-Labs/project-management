package model

import (
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"gorm.io/datatypes"
)

// Task is the GORM persistence shape for domain.Task (columns only).
type Task struct {
	ID            string          `gorm:"primaryKey"`
	Title         string          `gorm:"not null"`
	InitialPrompt string          `gorm:"type:text;not null"`
	Status        domain.Status   `gorm:"not null;index;check:chk_tasks_status,status IN ('ready','running','blocked','review','pr_ready','done','failed','on_hold','closed')"`
	Priority      domain.Priority `gorm:"not null;check:chk_tasks_priority,priority IN ('low','medium','high','critical')"`
	ProjectID     *string         `gorm:"index:idx_tasks_project_number,unique,priority:1;index"`
	// RepositoryID is the registered repo for managed-worktree allocate.
	RepositoryID *string `gorm:"index"`
	// Number is the per-project sequential ref (#N). Unique with project_id when both set.
	Number *int `gorm:"index:idx_tasks_project_number,unique,priority:2"`
	// JSONSlice (not serializer:json): empty slices must persist as "[]". GORM's
	// serializer:json writes "" for empty []string, which Postgres rejects on jsonb.
	Tags                datatypes.JSONSlice[string] `gorm:"column:tags;type:jsonb;not null;default:'[]'"`
	Milestone           *string                     `gorm:"index"`
	Gate                *domain.TaskGate            `gorm:"column:gate;serializer:json;type:jsonb"`
	Runner              string                      `gorm:"not null;default:'cursor'"`
	CursorModel         string                      `gorm:"not null;default:''"`
	RunnerConfig        datatypes.JSON              `gorm:"column:runner_config;type:jsonb;not null;default:'{}'"`
	PickupNotBefore     *time.Time                  `gorm:"index"`
	CriteriaSatisfiedAt *time.Time                  `gorm:"index"`
	PullRequestURL      *string                     `gorm:"type:text"`
	PendingRetry        *domain.PendingRetry        `gorm:"column:pending_retry;serializer:json;type:jsonb"`
	WorktreeID          *string                     `gorm:"index"`
}

// TableName pins the tasks table name.
func (Task) TableName() string { return "tasks" }
