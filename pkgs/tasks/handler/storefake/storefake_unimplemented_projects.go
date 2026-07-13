package storefake

import (
	"context"

	projectscontract "github.com/AlexsanderHamir/Hamix/pkgs/projects/contract"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
)

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateProject(context.Context, projectscontract.CreateProjectInput) (projectsdomain.Project, error) {
	return projectsdomain.Project{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjects(context.Context, bool, int) ([]projectsdomain.Project, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetProject(context.Context, string) (projectsdomain.Project, error) {
	return projectsdomain.Project{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateProject(context.Context, string, projectscontract.UpdateProjectInput) (projectsdomain.Project, error) {
	return projectsdomain.Project{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteProject(context.Context, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateProjectContext(context.Context, string, projectscontract.CreateProjectContextInput) (projectsdomain.ProjectContextItem, error) {
	return projectsdomain.ProjectContextItem{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjectContext(context.Context, string, bool, int) ([]projectsdomain.ProjectContextItem, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjectContextEdges(context.Context, string, []string) ([]projectsdomain.ProjectContextEdge, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateProjectContextEdge(context.Context, string, projectscontract.CreateProjectContextEdgeInput) (projectsdomain.ProjectContextEdge, error) {
	return projectsdomain.ProjectContextEdge{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateProjectContextEdge(context.Context, string, string, projectscontract.UpdateProjectContextEdgeInput) (projectsdomain.ProjectContextEdge, error) {
	return projectsdomain.ProjectContextEdge{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteProjectContextEdge(context.Context, string, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateProjectContext(context.Context, string, string, projectscontract.UpdateProjectContextInput) (projectsdomain.ProjectContextItem, error) {
	return projectsdomain.ProjectContextItem{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteProjectContext(context.Context, string, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjectsByRepository(context.Context, string) ([]projectsdomain.Project, error) {
	return nil, errNotImplemented
}
