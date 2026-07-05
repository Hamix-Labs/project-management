package model

import (
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// Project is the GORM persistence shape for domain.Project (columns only).
type Project struct {
	ID             string               `gorm:"primaryKey"`
	Name           string               `gorm:"not null;index"`
	Description    string               `gorm:"type:text;not null;default:''"`
	Status         domain.ProjectStatus `gorm:"not null;index;default:active;check:chk_projects_status,status IN ('active','archived')"`
	ContextSummary string               `gorm:"type:text;not null;default:''"`
	RepositoryID   *string              `gorm:"index"`
	IsDefault      bool                 `gorm:"not null;default:false;index"`
	CreatedAt      time.Time            `gorm:"not null;index"`
	UpdatedAt      time.Time            `gorm:"not null;index"`
}

// TableName pins the projects table name.
func (Project) TableName() string { return "projects" }
