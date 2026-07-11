package contract

import taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"

type TaskCRUDStore = taskcorecontract.TaskCRUDStore

type (
	CreateTaskInput      = taskcorecontract.CreateTaskInput
	UpdateTaskInput      = taskcorecontract.UpdateTaskInput
	ListFilter           = taskcorecontract.ListFilter
	ProjectFieldPatch    = taskcorecontract.ProjectFieldPatch
	PickupNotBeforePatch = taskcorecontract.PickupNotBeforePatch
	RequestRetryInput    = taskcorecontract.RequestRetryInput
)
