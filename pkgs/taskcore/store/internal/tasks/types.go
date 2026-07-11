package tasks

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
)

// CreateInput is the task creation payload. Re-aliased by the public
// store facade as store.CreateTaskInput so handler code stays
// unchanged.
type CreateInput = contract.CreateTaskInput

// PickupNotBeforePatch updates pickup_not_before when non-nil. Clear true means
// set the column to NULL (the task is no longer scheduled). Re-aliased by the
// public store facade as store.PickupNotBeforePatch. See docs/data-model.md.
type PickupNotBeforePatch = contract.PickupNotBeforePatch

// UpdateInput is the task patch payload. Each pointer field is
// optional; a nil pointer means "do not change". Re-aliased by the
// public store facade as store.UpdateTaskInput.
type UpdateInput = contract.UpdateTaskInput

// ListFilter optionally restricts flat task listing.
type ListFilter = contract.ListFilter

// ProjectFieldPatch updates project_id when non-nil. Clear true means
// set project_id to null. Re-aliased by the public store facade as
// store.ProjectFieldPatch.
type ProjectFieldPatch = contract.ProjectFieldPatch
