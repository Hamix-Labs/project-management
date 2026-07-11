package handler

import (
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
)

type projectsListResponse struct {
	Projects []projectsdomain.Project `json:"projects"`
	Limit    int                      `json:"limit"`
}
