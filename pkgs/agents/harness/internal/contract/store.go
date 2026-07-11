package contract

// Store is the persistence contract for harness orchestration, verify,
// resume, and git subpackages.
type Store interface {
	TaskStore
	CycleStore
	ChecklistStore
	SnapshotStore
	SettingsStore
	EventStore
	ProjectStore
}
