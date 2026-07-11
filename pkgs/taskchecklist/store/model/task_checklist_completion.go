package model

import (
	"time"

	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
)

// TaskChecklistCompletion is the GORM persistence shape for domain.TaskChecklistCompletion.
type TaskChecklistCompletion struct {
	TaskID            string                       `gorm:"primaryKey"`
	ItemID            string                       `gorm:"primaryKey"`
	At                time.Time                    `gorm:"not null"`
	By                string                       `gorm:"column:done_by;not null"`
	Evidence          string                       `gorm:"not null;default:'';type:text"`
	VerifiedBy        checklistdomain.VerifierKind `gorm:"not null;default:''"`
	VerifierReasoning string                       `gorm:"not null;default:'';type:text"`
	CycleID           string                       `gorm:"not null;default:''"`
}

// TableName pins the task_checklist_completions table name.
func (TaskChecklistCompletion) TableName() string { return "task_checklist_completions" }
