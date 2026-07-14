package realtime

// NotifyChangeFunc publishes a hint-with-id SSE frame after a mutation
// (task or project). Implementations live on the composition handler;
// BC handlers accept this type so composition can wire hub.Publish.
type NotifyChangeFunc func(typ ChangeType, id string)

// NotifyFunc is the project-handler name for the same hint-with-id signature.
type NotifyFunc = NotifyChangeFunc

// NotifyTaskChangedFunc publishes an enriched task SSE frame with optional data.
type NotifyTaskChangedFunc func(typ ChangeType, id string, data any)
