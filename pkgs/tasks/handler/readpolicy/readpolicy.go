// Package readpolicy holds shared read-side limits for handler aggregates.
// Constants mirror the SPA query policy in web/src/lib/readLimits.ts and
// useTasksApp so bootstrap and optional shell routes stay aligned with the
// client without importing HTTP or database packages.
package readpolicy

import taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"

const (
	// BootstrapListLimit matches the SPA home-page task list (limit 20).
	BootstrapListLimit = 20

	// BootstrapProjectsLimit matches AppShell useProjects (limit 100).
	BootstrapProjectsLimit = 100

	// BootstrapDraftsLimit matches useTaskCreateFlow drafts query.
	BootstrapDraftsLimit = 50

	// ShellChecklistIncluded documents that GET /v1/tasks/{id}/shell (when
	// shipped) embeds checklist items alongside the task row.
	ShellChecklistIncluded = true

	// TaskListDefaultLimit is GET /tasks default when ?limit= is omitted.
	// Owned by taskcore/contract; re-exported here for shell aggregates.
	TaskListDefaultLimit = taskcorecontract.TaskListDefaultLimit

	// TaskListMaxLimit caps GET /tasks ?limit=.
	TaskListMaxLimit = taskcorecontract.TaskListMaxLimit

	// TaskEventsDefaultLimit is GET /tasks/{id}/events default page size.
	TaskEventsDefaultLimit = 50

	// TaskEventsMaxLimit caps GET /tasks/{id}/events ?limit=.
	TaskEventsMaxLimit = 200

	// CycleListDefaultLimit is GET /tasks/{id}/cycles default page size.
	CycleListDefaultLimit = 50

	// CycleListMaxLimit caps GET /tasks/{id}/cycles ?limit=.
	CycleListMaxLimit = 200

	// CycleStreamDefaultLimit is GET /tasks/{id}/cycles/{cycleId}/stream default.
	CycleStreamDefaultLimit = 100

	// CycleStreamMaxLimit caps cycle stream ?limit=.
	CycleStreamMaxLimit = 500
)
