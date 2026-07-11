package handler

import taskcorehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"

type (
	listResponse              = taskcorehandler.ListResponse
	taskStatsResponse         = taskcorehandler.TaskStatsResponse
	dependsOnWire             = taskcorehandler.DependsOnWire
	taskComposePayloadJSON    = taskcorehandler.TaskComposePayloadJSON
	patchPickupNotBeforeField = taskcorehandler.PatchPickupNotBeforeField
	taskCreateJSON            = taskcorehandler.TaskCreateJSON
	taskPatchJSON             = taskcorehandler.TaskPatchJSON
)

var (
	buildListResponse          = taskcorehandler.BuildListResponse
	taskStatsResponseFromStore = taskcorehandler.TaskStatsResponseFromStore
	decodeComposePayload       = taskcorehandler.DecodeComposePayload
	composePayloadToRaw        = taskcorehandler.ComposePayloadToRaw
)

const (
	maxListIntQueryParamBytes = taskcorehandler.MaxListIntQueryParamBytes
	maxListAfterIDParamBytes  = taskcorehandler.MaxListAfterIDParamBytes
)

type taskDependenciesListResponse = taskcorehandler.TaskDependenciesListResponse

var parseListParams = taskcorehandler.ParseListParams

var pickupNotBeforeMinAllowed = taskcorehandler.PickupNotBeforeMinAllowed

var (
	taskapiDomainTasksCreatedTotal = taskcorehandler.DomainTasksCreatedTotal
	taskapiDomainTasksUpdatedTotal = taskcorehandler.DomainTasksUpdatedTotal
	taskapiDomainTasksDeletedTotal = taskcorehandler.DomainTasksDeletedTotal
)
