package store

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

// HandlerAPI is the persistence contract required by pkgs/tasks/handler.
// *Store implements it; compile-time checks live in handler/handler_store_assert_test.go.
type HandlerAPI interface {
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
