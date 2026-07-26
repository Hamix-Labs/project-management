package contract

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
)

// CreateProjectInput is the store input for creating a project.
type CreateProjectInput struct {
	ID           string
	Name         string
	Description  string
	RepositoryID *string
}

// UpdateProjectInput is a partial patch for project metadata.
type UpdateProjectInput struct {
	Name        *string
	Description *string
	Status      *domain.ProjectStatus
}
