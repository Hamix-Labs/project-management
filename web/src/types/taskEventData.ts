/**
 * Narrowed `TaskEvent.data` shapes for high-traffic audit event families.
 * Long-tail types keep `Record<string, unknown>` via the TaskEvent union.
 */

export type PhaseLifecycleEventData = {
  phase?: string;
  status?: string;
  summary?: string;
  cycle_id?: string;
  phase_seq?: number;
  details?: Record<string, unknown>;
};

export type CycleLifecycleEventData = {
  cycle_id?: string;
  attempt_seq?: number;
  status?: string;
  reason?: string;
  failure_summary?: string;
};

export type TransitionEventData = {
  from?: string;
  to?: string;
};

export type ChecklistItemRemovedEventData = {
  text?: string;
};

export type TaskRetryRequestedEventData = {
  mode?: string;
};

export type TaskPickupFailedEventData = {
  reason?: string;
};

export type PhaseLifecycleEventType =
  | "phase_started"
  | "phase_completed"
  | "phase_failed"
  | "phase_skipped";

export type CycleLifecycleEventType =
  | "cycle_started"
  | "cycle_completed"
  | "cycle_failed";

export type TransitionEventType =
  | "status_changed"
  | "priority_changed"
  | "message_added"
  | "prompt_appended";

export type TypedTaskEventDataType =
  | PhaseLifecycleEventType
  | CycleLifecycleEventType
  | TransitionEventType
  | "checklist_item_removed"
  | "task_retry_requested"
  | "task_pickup_failed";
