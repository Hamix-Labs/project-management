package contract

// HandlerStore is the composed persistence contract required by pkgs/tasks/handler.
type HandlerStore interface {
	HealthStore
	SettingsStore
	TaskCRUDStore
	TaskEventStore
	ChecklistStore
	CycleStore
	ProjectStore
	DraftStore
	TemplateStore
	GitReadStore
	GitWriteStore
}
