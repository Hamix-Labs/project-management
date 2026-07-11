package domain

import (
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

type (
	// Actor aliases taskcore/domain (canonical owner).
	Actor = taskcoredomain.Actor
	// TaskEvent aliases taskevents/domain.
	TaskEvent = taskeventsdomain.TaskEvent
	// EventType aliases taskevents/domain.
	EventType = taskeventsdomain.EventType
	// ResponseThreadEntry aliases taskevents/domain.
	ResponseThreadEntry = taskeventsdomain.ResponseThreadEntry
)

const (
	ActorUser  = taskcoredomain.ActorUser
	ActorAgent = taskcoredomain.ActorAgent

	EventTaskCreated           = taskeventsdomain.EventTaskCreated
	EventStatusChanged         = taskeventsdomain.EventStatusChanged
	EventPriorityChanged       = taskeventsdomain.EventPriorityChanged
	EventPromptAppended        = taskeventsdomain.EventPromptAppended
	EventContextAdded          = taskeventsdomain.EventContextAdded
	EventConstraintAdded       = taskeventsdomain.EventConstraintAdded
	EventSuccessCriterionAdded = taskeventsdomain.EventSuccessCriterionAdded
	EventNonGoalAdded          = taskeventsdomain.EventNonGoalAdded
	EventPlanAdded             = taskeventsdomain.EventPlanAdded
	EventChecklistItemAdded    = taskeventsdomain.EventChecklistItemAdded
	EventChecklistItemToggled  = taskeventsdomain.EventChecklistItemToggled
	EventChecklistItemUpdated  = taskeventsdomain.EventChecklistItemUpdated
	EventChecklistItemRemoved  = taskeventsdomain.EventChecklistItemRemoved
	EventMessageAdded          = taskeventsdomain.EventMessageAdded
	EventArtifactAdded         = taskeventsdomain.EventArtifactAdded
	EventApprovalRequested     = taskeventsdomain.EventApprovalRequested
	EventApprovalGranted       = taskeventsdomain.EventApprovalGranted
	EventTaskCompleted         = taskeventsdomain.EventTaskCompleted
	EventOnTaskDone            = taskeventsdomain.EventOnTaskDone
	EventTaskFailed            = taskeventsdomain.EventTaskFailed
	EventTaskRetryRequested    = taskeventsdomain.EventTaskRetryRequested
	EventTaskPickupFailed      = taskeventsdomain.EventTaskPickupFailed
	EventCycleStarted          = taskeventsdomain.EventCycleStarted
	EventCycleCompleted        = taskeventsdomain.EventCycleCompleted
	EventCycleFailed           = taskeventsdomain.EventCycleFailed
	EventPhaseStarted          = taskeventsdomain.EventPhaseStarted
	EventPhaseCompleted        = taskeventsdomain.EventPhaseCompleted
	EventPhaseFailed           = taskeventsdomain.EventPhaseFailed
	EventPhaseSkipped          = taskeventsdomain.EventPhaseSkipped
	EventSyncPing              = taskeventsdomain.EventSyncPing
)

var EventTypeAcceptsUserResponse = taskeventsdomain.EventTypeAcceptsUserResponse
