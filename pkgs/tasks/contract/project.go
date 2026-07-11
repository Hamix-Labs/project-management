package contract

import projectscontract "github.com/AlexsanderHamir/Hamix/pkgs/projects/contract"

// ProjectStore covers projects and the project context graph.
type ProjectStore = projectscontract.ProjectStore

// CreateProjectInput is the store input for creating a project.
type CreateProjectInput = projectscontract.CreateProjectInput

// UpdateProjectInput is a partial patch for project metadata.
type UpdateProjectInput = projectscontract.UpdateProjectInput

// CreateProjectContextInput is the store input for appending a project context item.
type CreateProjectContextInput = projectscontract.CreateProjectContextInput

// UpdateProjectContextInput is a partial patch for one project context item.
type UpdateProjectContextInput = projectscontract.UpdateProjectContextInput

// CreateProjectContextEdgeInput is the store input for connecting two context nodes.
type CreateProjectContextEdgeInput = projectscontract.CreateProjectContextEdgeInput

// UpdateProjectContextEdgeInput is a partial patch for one project context edge.
type UpdateProjectContextEdgeInput = projectscontract.UpdateProjectContextEdgeInput
