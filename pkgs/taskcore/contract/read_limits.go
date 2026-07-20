package contract

// Task list pagination limits for GET /tasks.
// Keep aligned with pkgs/tasks/handler/readpolicy and web/src/lib/readLimits.ts.
const (
	// TaskListDefaultLimit is GET /tasks default when ?limit= is omitted.
	TaskListDefaultLimit = 50

	// TaskListMaxLimit caps GET /tasks ?limit=.
	TaskListMaxLimit = 200
)
