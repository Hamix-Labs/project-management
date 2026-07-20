package composition

import (
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/storehooks"
	gitinventorystore "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store"
	projectsstore "github.com/AlexsanderHamir/Hamix/pkgs/projects/store"
	settingsstore "github.com/AlexsanderHamir/Hamix/pkgs/settings/store"
	checkliststore "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/store"
	composestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/store"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	taskeventsstore "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store"
	"gorm.io/gorm"
)

// API is the taskapi composition root: BC stores plus cross-cutting hooks.
// It satisfies pkgs/tasks/wire.HandlerAPI and pkgs/agents/worker.Store.
type API struct {
	taskcore  *taskcorestore.Store
	projects  *projectsstore.Store
	git       *gitinventorystore.Store
	settings  *settingsstore.Store
	compose   *composestore.Store
	checklist *checkliststore.Store
	cycles    *cyclesstore.Store
	events    *taskeventsstore.Store
	hooks     *storehooks.Runtime
}

// NewAPI constructs BC stores sharing db.
//
//funclogmeasure:skip category=hot-path reason="Composition wiring; BC store constructors emit operation traces."
func NewAPI(db *gorm.DB) *API {
	git := gitinventorystore.NewStore(db, nil)
	projects := projectsstore.NewStore(db)
	return &API{
		taskcore:  taskcorestore.NewStore(db),
		projects:  projects,
		git:       git,
		settings:  settingsstore.NewStore(db),
		compose:   composestore.NewStore(db),
		checklist: checkliststore.NewStore(db),
		cycles:    cyclesstore.NewStore(db),
		events:    taskeventsstore.NewStore(db),
		hooks:     storehooks.NewRuntime(),
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by storehooks chokepoints."
func (a *API) gitDeps() storehooks.GitDeps {
	if a == nil {
		return storehooks.GitDeps{}
	}
	return storehooks.GitDeps{Git: a.git, Projects: a.projects}
}

// SetReadyTaskNotifier registers n for ready-task notifications.
//
//funclogmeasure:skip category=hot-path reason="Startup wiring hook; notify path traces at scheduling chokepoints."
func (a *API) SetReadyTaskNotifier(n storehooks.ReadyTaskNotifier) {
	if a == nil {
		return
	}
	a.hooks.SetReadyTaskNotifier(n)
}

// SetPickupWake registers w for deferred-pickup scheduling.
//
//funclogmeasure:skip category=hot-path reason="Startup wiring hook; pickup wake traces at scheduling chokepoints."
func (a *API) SetPickupWake(w storehooks.PickupWake) {
	if a == nil {
		return
	}
	a.hooks.SetPickupWake(w)
}

// TaskCore returns the taskcore store (for tests that need direct BC access).
//
//funclogmeasure:skip category=hot-path reason="Pure accessor; BC store methods emit operation traces."
func (a *API) TaskCore() *taskcorestore.Store {
	if a == nil {
		return nil
	}
	return a.taskcore
}

// GitStore returns the git inventory store.
//
//funclogmeasure:skip category=hot-path reason="Pure accessor; BC store methods emit operation traces."
func (a *API) GitStore() *gitinventorystore.Store {
	if a == nil {
		return nil
	}
	return a.git
}
