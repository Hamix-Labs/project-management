package contract

// ChangeType names SSE payload types that taskcore handlers emit after mutations.
// Wire values must stay aligned with pkgs/tasks/realtime (ADR-0035 owner).
type ChangeType string

const (
	ChangeTaskCreated           ChangeType = "task_created"
	ChangeTaskUpdated           ChangeType = "task_updated"
	ChangeTaskDeleted           ChangeType = "task_deleted"
	ChangeTaskGateChanged       ChangeType = "task_gate_changed"
	ChangeTaskDependencyChanged ChangeType = "task_dependency_changed"
)

// NotifyChangeFunc publishes a hint-with-id SSE frame after a mutation.
// Composition wires this to the realtime hub; taskcore never imports realtime.
type NotifyChangeFunc func(typ ChangeType, id string)

// NotifyTaskChangedFunc publishes an enriched task SSE frame with optional data.
type NotifyTaskChangedFunc func(typ ChangeType, id string, data any)
