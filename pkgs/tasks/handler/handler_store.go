package handler

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

// HandlerStore is the composed persistence contract required by pkgs/tasks/handler.
// Slices live in bounded-context contract packages; this type composes them at the
// taskapi wiring edge only.
type HandlerStore interface {
	taskcorecontract.HealthStore
	settingscontract.SettingsStore
	taskcorecontract.TaskCRUDStore
	taskeventscontract.TaskEventStore
	checklistcontract.ChecklistStore
	cyclescontract.CycleStore
	projectscontract.ProjectStore
	composecontract.ComposeStore
	gitcontract.GitReadStore
	gitcontract.GitWriteStore
}
