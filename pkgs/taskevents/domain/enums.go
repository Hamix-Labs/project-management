package domain

type EventType string

const (
	EventTaskCreated           EventType = "task_created"
	EventStatusChanged         EventType = "status_changed"
	EventPriorityChanged       EventType = "priority_changed"
	EventPromptAppended        EventType = "prompt_appended"
	EventContextAdded          EventType = "context_added"
	EventConstraintAdded       EventType = "constraint_added"
	EventSuccessCriterionAdded EventType = "success_criterion_added"
	EventNonGoalAdded          EventType = "non_goal_added"
	EventPlanAdded             EventType = "plan_added"
	EventChecklistItemAdded    EventType = "checklist_item_added"
	EventChecklistItemToggled  EventType = "checklist_item_toggled"
	EventChecklistItemUpdated  EventType = "checklist_item_updated"
	EventChecklistItemRemoved  EventType = "checklist_item_removed"
	EventMessageAdded          EventType = "message_added"
	EventArtifactAdded         EventType = "artifact_added"
	EventApprovalRequested     EventType = "approval_requested"
	EventApprovalGranted       EventType = "approval_granted"
	EventTaskCompleted         EventType = "task_completed"
	EventOnTaskDone            EventType = "on_task_done"
	EventTaskFailed            EventType = "task_failed"
	EventTaskRetryRequested    EventType = "task_retry_requested"
	EventTaskPolishRequested   EventType = "task_polish_requested"
	// EventTaskPickupFailed is emitted when the worker cannot persist ready→running
	// (e.g. store/jsonb errors). The task stays ready; pickup is deferred.
	EventTaskPickupFailed EventType = "task_pickup_failed"
	// Execution-cycle audit mirrors. Emitted in the same SQL transaction as writes to
	// task_cycles / task_cycle_phases so GET /tasks/{id}/events stays a complete witness
	// of cycle activity. See docs/data-model.md.
	EventCycleStarted   EventType = "cycle_started"
	EventCycleCompleted EventType = "cycle_completed"
	EventCycleFailed    EventType = "cycle_failed"
	EventPhaseStarted   EventType = "phase_started"
	EventPhaseCompleted EventType = "phase_completed"
	EventPhaseFailed    EventType = "phase_failed"
	EventPhaseSkipped   EventType = "phase_skipped"
	// EventSyncPing is included in the dev ticker rotation (HAMIX_SSE_TEST) alongside every other EventType.
	EventSyncPing EventType = "sync_ping"
)
