// Package wire holds taskapi wiring-edge composed contracts. Types here compose
// bounded-context contract slices for handler and store assembly only — not domain logic.
package wire

import (
	gitcontract "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	projectscontract "github.com/AlexsanderHamir/Hamix/pkgs/projects/contract"
	settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	composecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/contract"
	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	taskeventscontract "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
)

// HandlerAPI is the composed persistence contract required by pkgs/tasks/handler
// and implemented by *internal/taskapi/composition.API at the taskapi wiring edge.
type HandlerAPI interface {
	taskcorecontract.HealthStore
	settingscontract.SettingsStore
	taskcorecontract.TaskCRUDStore
	taskeventscontract.TaskEventStore
	checklistcontract.ChecklistStore
	cyclescontract.CycleStore
	cyclescontract.CycleFailuresStore
	projectscontract.ProjectStore
	composecontract.ComposeStore
	gitcontract.GitInventoryStore
	gitcontract.GitWriteStore
}
