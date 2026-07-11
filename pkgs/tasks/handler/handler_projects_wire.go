package handler

import (
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
)

// projectsListResponse mirrors GET /projects and bootstrap projects payload.
type projectsListResponse struct {
	Projects []projectsdomain.Project `json:"projects"`
	Limit    int                      `json:"limit"`
}
