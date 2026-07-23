package storehooks

import (
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/storehooks/notify"
)

// Runtime groups cross-cutting hooks registered once at taskapi startup.
type Runtime struct {
	Notify            notify.Holder
	PickupWake        PickupWakeRegistry
	WorktreeProvision WorktreeProvisionRegistry
}

// NewRuntime returns a ready-to-wire hook bundle.
//
//funclogmeasure:skip category=hot-path reason="Composition wiring; hook users emit operation traces."
func NewRuntime() *Runtime {
	return &Runtime{}
}

// ReadyTaskNotifier is the notifier type wired by cmd/taskapi at startup.
type ReadyTaskNotifier = notify.Notifier

// SetReadyTaskNotifier registers n for ready-task notifications.
//
//funclogmeasure:skip category=hot-path reason="Startup wiring hook; notify path traces at scheduling chokepoints."
func (r *Runtime) SetReadyTaskNotifier(n ReadyTaskNotifier) {
	if r == nil {
		return
	}
	r.Notify.Set(n)
}

// SetPickupWake registers w for deferred-pickup scheduling.
//
//funclogmeasure:skip category=hot-path reason="Startup wiring hook; pickup wake traces at scheduling chokepoints."
func (r *Runtime) SetPickupWake(w PickupWake) {
	if r == nil {
		return
	}
	r.PickupWake.Set(w)
}

// SetWorktreeProvisioner registers p for async managed-worktree allocate after create.
//
//funclogmeasure:skip category=hot-path reason="Startup wiring hook; provision traces at composition chokepoints."
func (r *Runtime) SetWorktreeProvisioner(p WorktreeProvisioner) {
	if r == nil {
		return
	}
	r.WorktreeProvision.Set(p)
}
