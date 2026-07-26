package handler

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
)

type projectCreateJSON struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	RepositoryID *string `json:"repository_id"`
}

type projectPatchJSON struct {
	Name        *string               `json:"name"`
	Description *string               `json:"description"`
	Status      *domain.ProjectStatus `json:"status"`
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p projectPatchJSON) isEmpty() bool {
	return p.Name == nil && p.Description == nil && p.Status == nil
}

type projectsListResponse struct {
	Projects []domain.Project `json:"projects"`
	Limit    int              `json:"limit"`
}
