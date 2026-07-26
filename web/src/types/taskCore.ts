export type { GateCriterion, GateStatus, TaskGate } from "./gate";
import type { TaskGate } from "./gate";

export type Status =
  | "ready"
  | "running"
  | "blocked"
  | "review"
  | "done"
  | "failed"
  | "on_hold"
  | "closed";

export type Priority = "low" | "medium" | "high" | "critical";

/** Empty string means no selection yet (create / draft forms). */
export type PriorityChoice = Priority | "";

export type TaskDependencySatisfies = "done";

export type TaskDependencyEdge = {
  task_id: string;
  satisfies?: TaskDependencySatisfies;
};

export type Task = {
  id: string;
  /**
   * Per-project sequential number surfaced in the SPA as `#N`. Server
   * assigns on create for tasks that belong to a project (see
   * docs/data-model.md). Absent (or `null`) for legacy tasks and for
   * global tasks that were created before the numbering migration ran.
   */
  number?: number | null;
  title: string;
  initial_prompt: string;
  status: Status;
  priority: Priority;
  /** Agent runner id for this task (e.g. `cursor`); chosen at create time. */
  runner: string;
  /**
   * Optional `cursor-agent --model` value for this task. Empty means omit
   * the flag (Cursor default for the account).
   */
  cursor_model: string;
  /**
   * Optional verify chat policy override (`same_chat` | `different_chat`).
   * Empty / omitted inherits `app_settings.verify_chat_mode`.
   */
  verify_chat_mode?: string;
  /**
   * When set (RFC3339), the agent worker will not dequeue this ready task
   * until this instant. Omitted when eligible immediately.
   */
  pickup_not_before?: string;
  /** RFC3339 UTC from the task_created audit event; present on list/get responses. */
  created_at?: string;
  /** Present when this task belongs to a long-lived project context. */
  project_id?: string;
  /** Git worktree binding (ADR-0039). */
  worktree_id?: string;
  tags?: string[];
  milestone?: string | null;
  depends_on?: TaskDependencyEdge[];
  criteria_satisfied_at?: string;
  gate?: TaskGate | null;
};

export type TaskListResponse = {
  tasks: Task[];
  limit: number;
  offset: number;
  /** True when the server may have more tasks (see GET /tasks in docs/api.md). */
  has_more: boolean;
};

/**
 * One entry in `recent_failures` from `GET /tasks/stats`. Mirrors the
 * server projection in pkgs/tasks/handler/handler_task_json.go (struct
 * taskStatsFailureJSON). `task_id` + `event_seq` deep-link to
 * `GET /tasks/{task_id}/events/{event_seq}`; `status` is the original
 * terminal cycle status (`failed` or `aborted`) recovered from the
 * cycle_failed payload (the mirror folds aborts into cycle_failed).
 */
export type TaskStatsRecentFailure = {
  task_id: string;
  event_seq: number;
  /** ISO 8601 from API. */
  at: string;
  cycle_id: string;
  attempt_seq: number;
  status: "failed" | "aborted";
  /**
   * Human-readable failure text: prefers `failure_summary` on the
   * cycle_failed mirror (same source as execute phase_failed), then
   * legacy enrichment from a matching phase_failed event, else the
   * mirror reason code (e.g. runner_non_zero_exit).
   */
  reason: string;
};

/** Fixed set of event types surfaced by GET /tasks/activity. */
export type TaskActivityEventType =
  | "status_changed"
  | "phase_failed"
  | "approval_granted";

/** One row from GET /tasks/activity. */
export type TaskActivityEvent = {
  task_id: string;
  /** 1-based audit-trail sequence number on the owning task. */
  seq: number;
  /** ISO 8601 UTC timestamp. */
  at: string;
  type: TaskActivityEventType;
  by: string;
  data: Record<string, unknown>;
  task_title?: string;
  /**
   * Per-project sequential task number surfaced by the server when the
   * owning task has one. When present the SPA renders `#N` on the
   * timeline; otherwise it falls back to the shortened UUID.
   */
  task_number?: number | null;
  /** Joined from the owning task — Timeline priority filter. */
  task_priority?: Priority;
  /** Joined from the owning task — Timeline project filter. */
  task_project_id?: string | null;
  /** Joined from the owning task — Timeline tag filter. */
  task_tags?: string[];
};

/** Response envelope for GET /tasks/activity. */
export type TaskActivityResponse = {
  total: number;
  limit: number;
  offset: number;
  events: TaskActivityEvent[];
};

/** GET /tasks/cycle-failures — paginated cycle_failed list for the failures page. */
export type CycleFailuresListResponse = {
  total: number;
  limit: number;
  offset: number;
  sort: string;
  reason_sort_truncated: boolean;
  failures: TaskStatsRecentFailure[];
};

/**
 * Cycle aggregates from `GET /tasks/stats`. Both maps are always
 * present (`{}` on empty database). Inner enums match
 * `pkgs/tasks/domain` exactly so a future enum change trips the
 * parser, contract tests, and any phase/status view in the same PR.
 */
export type TaskStatsCycles = {
  by_status: Partial<Record<import("./cycle").CycleStatus, number>>;
  by_triggered_by: Partial<Record<"user" | "agent", number>>;
};

/**
 * Phase aggregates from `GET /tasks/stats`. The outer map keys are the
 * writable `domain.Phase` values (`execute`, `verify`); every key is
 * always present (the inner map is `{}` for phases that have never
 * run). Legacy phase buckets returned by historical task_cycle_phases
 * rows are dropped at the parser boundary — see
 * {@link import("./cycle").WritablePhase}.
 */
export type TaskStatsPhases = {
  by_phase_status: Record<
    import("./cycle").WritablePhase,
    Partial<Record<import("./cycle").PhaseStatus, number>>
  >;
};

/**
 * One entry in the runner / model breakdown returned by
 * `GET /tasks/stats` (Phase 2 of the per-task runner+model
 * attribution work). `succeeded` mirrors `by_status.succeeded` so
 * the SPA can branch on the percentile gate without a missing-key
 * check; `duration_p50_succeeded_seconds` /
 * `duration_p95_succeeded_seconds` are 0 when `succeeded === 0`
 * (render "—" rather than "0.00s" in that case).
 */
export type TaskStatsRunnerBucket = {
  by_status: Partial<Record<import("./cycle").CycleStatus, number>>;
  succeeded: number;
  duration_p50_succeeded_seconds: number;
  duration_p95_succeeded_seconds: number;
};

/**
 * Per-runner / per-model / per-(runner,model) aggregates on
 * `GET /tasks/stats`. All three maps are always present (`{}` on
 * empty database). Bucket keys are verbatim from cycle meta:
 *  - `by_runner` is keyed by `runner.Name()` ("unknown" for cycles
 *    whose meta predates the V2 keys)
 *  - `by_model` is keyed by the resolved effective model; the
 *    empty-string key is preserved (means "no model configured")
 *  - `by_runner_model` is the (runner, model) pair joined by `|`
 *  - `by_runner_model_resolved` is the (runner, effective model,
 *    resolved model) triple joined by `|`. Only populated for cycles
 *    whose execute-phase details surfaced a concrete resolved model
 *    (today: cursor-agent's `system.init.model` event under
 *    `--output-format stream-json`). Cycles without a resolved model
 *    are intentionally absent so the SPA only renders a "→ actual
 *    model" sub-row when we have a real observation, not a guess.
 */
export type TaskStatsRunner = {
  by_runner: Record<string, TaskStatsRunnerBucket>;
  by_model: Record<string, TaskStatsRunnerBucket>;
  by_runner_model: Record<string, TaskStatsRunnerBucket>;
  by_runner_model_resolved: Record<string, TaskStatsRunnerBucket>;
};

export type TaskStatsResponse = {
  total: number;
  ready: number;
  critical: number;
  /**
   * Count of `status='ready'` tasks intentionally deferred via
   * `pickup_not_before > now()`. Always present (`0` on a fresh
   * database). Stats consumers use this to distinguish
   * "0 ready, 12 scheduled" (intentionally deferred — agent worker is
   * correctly idle) from "0 ready, 0 scheduled" (truly idle, nothing
   * to do). Defaults to `0` when an older backend omits the key
   * (parser sets it explicitly so callers can rely on a number).
   */
  scheduled: number;
  by_status: Partial<Record<Status, number>>;
  by_priority: Partial<Record<Priority, number>>;
  cycles: TaskStatsCycles;
  phases: TaskStatsPhases;
  runner: TaskStatsRunner;
  /** Newest first; capped server-side at 25. Always an array (never null). */
  recent_failures: TaskStatsRecentFailure[];
};

export type TaskChangeType =
  | "task_created"
  | "task_updated"
  | "task_deleted"
  | "task_cycle_changed"
  | "project_created"
  | "project_updated"
  | "project_deleted"
;

/**
 * Wire shape of a single SSE frame on `GET /events` (legacy narrow subset).
 *
 * Full SSE surface is {@link SSEChangeType} + `TaskChangeFrame` in
 * `tasks/task-query/sseInvalidate.ts`. Prefer those for new code.
 *
 * `cycle_id` is only present on `task_cycle_changed` (omitted for the other
 * three types so the existing wire shape stays byte-identical).
 */
export type TaskChangeEvent = {
  type: TaskChangeType;
  id: string;
  cycle_id?: string;
};

/**
 * SSE `data.type` strings on `GET /events`. Mirrors
 * `pkgs/tasks/realtime/wire.go` (`realtime.ChangeType`).
 */
export const SSE_CHANGE_TYPE = {
  taskCreated: "task_created",
  taskUpdated: "task_updated",
  taskDeleted: "task_deleted",
  taskGateChanged: "task_gate_changed",
  taskDependencyChanged: "task_dependency_changed",
  taskCycleChanged: "task_cycle_changed",
  taskEventChanged: "task_event_changed",
  agentRunProgress: "agent_run_progress",
  projectCreated: "project_created",
  projectUpdated: "project_updated",
  projectDeleted: "project_deleted",
  settingsChanged: "settings_changed",
  agentRunCancelled: "agent_run_cancelled",
  resync: "resync",
} as const;

export const SSE_CHANGE_TYPES = Object.values(SSE_CHANGE_TYPE);

export type SSEChangeType = (typeof SSE_CHANGE_TYPE)[keyof typeof SSE_CHANGE_TYPE];

/** Task-scoped hint frames that carry a task `id` (invalidation path). */
export const SSE_TASK_HINT_TYPES = [
  SSE_CHANGE_TYPE.taskCreated,
  SSE_CHANGE_TYPE.taskUpdated,
  SSE_CHANGE_TYPE.taskDeleted,
  SSE_CHANGE_TYPE.taskGateChanged,
  SSE_CHANGE_TYPE.taskDependencyChanged,
] as const;

/** Project-scoped hint frames that carry a project `id`. */
export const SSE_PROJECT_HINT_TYPES = [
  SSE_CHANGE_TYPE.projectCreated,
  SSE_CHANGE_TYPE.projectUpdated,
  SSE_CHANGE_TYPE.projectDeleted,
] as const;

export const STATUSES: Status[] = [
  "ready",
  "running",
  "blocked",
  "review",
  "done",
  "failed",
  "on_hold",
  "closed",
];

/**
 * Status values operators may set via create/PATCH. `closed` is
 * intentionally excluded — it is reached only via `POST /tasks/{id}/close`
 * (see docs/data-model.md) so PATCH-status edits and create-form
 * status pickers never surface it as a writable choice.
 */
export const CLIENT_WRITABLE_STATUSES: Status[] = STATUSES.filter(
  (s) => s !== "closed",
);

export const PRIORITIES: Priority[] = [
  "low",
  "medium",
  "high",
  "critical",
];

/** New tasks start here; status is not user-selectable in the UI. */
export const DEFAULT_NEW_TASK_STATUS: Status = "ready";

/** Mirrors server `domain.EventType` (audit trail). */
export const TASK_EVENT_TYPES = [
  "task_created",
  "status_changed",
  "priority_changed",
  "prompt_appended",
  "context_added",
  "constraint_added",
  "success_criterion_added",
  "non_goal_added",
  "plan_added",
  "checklist_item_added",
  "checklist_item_toggled",
  "checklist_item_updated",
  "checklist_item_removed",
  "message_added",
  "artifact_added",
  "approval_requested",
  "approval_granted",
  "task_completed",
  /** Harness audit when task reaches done; no PR UI in v0.1 (see docs/data-model.md). */
  "on_task_done",
  "task_failed",
  "task_retry_requested",
  "task_polish_requested",
  "task_pickup_failed",
  // Execution-cycle audit mirrors. The backend writes these in the same
  // SQL transaction as task_cycles / task_cycle_phases rows so GET
  // /tasks/{id}/events is a complete witness of cycle activity (see
  // pkgs/tasks/domain/enums.go and docs/data-model.md). They land
  // on the timeline as soon as the agent worker dispatches a real task,
  // so omitting them from this allow-list makes parseTaskApi reject the
  // entire /events response with "event type must be a known value" and
  // collapses the Updates section into an error banner.
  "cycle_started",
  "cycle_completed",
  "cycle_failed",
  "phase_started",
  "phase_completed",
  "phase_failed",
  "phase_skipped",
  "sync_ping",
] as const;

export type TaskEventType = (typeof TASK_EVENT_TYPES)[number];

export type {
  ChecklistItemRemovedEventData,
  CycleLifecycleEventData,
  CycleLifecycleEventType,
  PhaseLifecycleEventData,
  PhaseLifecycleEventType,
  TaskPickupFailedEventData,
  TaskRetryRequestedEventData,
  TransitionEventData,
  TransitionEventType,
  TypedTaskEventDataType,
} from "./taskEventData";

import type {
  ChecklistItemRemovedEventData,
  CycleLifecycleEventData,
  CycleLifecycleEventType,
  PhaseLifecycleEventData,
  PhaseLifecycleEventType,
  TaskPickupFailedEventData,
  TaskRetryRequestedEventData,
  TransitionEventData,
  TransitionEventType,
  TypedTaskEventDataType,
} from "./taskEventData";

/** One message in the user ↔ agent thread on an event (`response_thread` in API). */
export type TaskEventResponseEntry = {
  /** ISO 8601 from API */
  at: string;
  by: "user" | "agent";
  body: string;
};

type TaskEventEnvelope = {
  seq: number;
  /** ISO 8601 from API */
  at: string;
  by: "user" | "agent";
  /** Human-submitted text for event types that accept input (`PATCH .../events/{seq}`). */
  user_response?: string;
  /** ISO 8601 when `user_response` was last saved; omitted for legacy rows. */
  user_response_at?: string;
  /** Ordered messages on this event (user and agent); legacy rows may be synthesized server-side. */
  response_thread?: TaskEventResponseEntry[];
};

/** High-traffic families carry narrowed `data`; long-tail types keep an opaque bag. */
export type TaskEvent =
  | (TaskEventEnvelope & {
      type: PhaseLifecycleEventType;
      data: PhaseLifecycleEventData;
    })
  | (TaskEventEnvelope & {
      type: CycleLifecycleEventType;
      data: CycleLifecycleEventData;
    })
  | (TaskEventEnvelope & {
      type: TransitionEventType;
      data: TransitionEventData;
    })
  | (TaskEventEnvelope & {
      type: "checklist_item_removed";
      data: ChecklistItemRemovedEventData;
    })
  | (TaskEventEnvelope & {
      type: "task_retry_requested";
      data: TaskRetryRequestedEventData;
    })
  | (TaskEventEnvelope & {
      type: "task_pickup_failed";
      data: TaskPickupFailedEventData;
    })
  | (TaskEventEnvelope & {
      type: Exclude<TaskEventType, TypedTaskEventDataType>;
      data: Record<string, unknown>;
    });

export type TaskEventsResponse = {
  task_id: string;
  events: TaskEvent[];
  /** From server when using paged `GET /tasks/{id}/events`; omitted on legacy full list. */
  limit?: number;
  total?: number;
  /** 1-based inclusive ranks in newest-first ordering (paged responses). */
  range_start?: number;
  range_end?: number;
  /** False when omitted in JSON (unpaged full list). */
  has_more_newer?: boolean;
  has_more_older?: boolean;
  /** Latest approval request still open (server-computed; not limited to the current page). */
  approval_pending: boolean;
};

/** Single row from `GET /tasks/{id}/events/{seq}` (same shape as one list element plus `task_id`). */
export type TaskEventDetail = TaskEvent & {
  task_id: string;
};

export type {
  ChecklistItemDraft,
  ChecklistVerifyCommandInput,
  TaskChecklistItemView,
  TaskChecklistResponse,
  TaskDraftChecklistItem,
} from "./checklist";
export {
  CHECKLIST_EVIDENCE_DISPLAY_CAP,
  DEFAULT_VERIFY_MAX_RETRIES,
} from "./checklist";
import type { TaskDraftChecklistItem } from "./checklist";

export type TaskComposePayload = {
  title: string;
  initial_prompt: string;
  status: Status;
  priority: Priority;
  runner?: string;
  cursor_model?: string;
  verify_chat_mode?: string;
  project_id?: string;
  repository_id?: string;
  worktree_id?: string;
  pickup_not_before?: string;
  tags?: string[];
  milestone?: string;
  depends_on?: TaskDependencyEdge[];
  checklist_items: TaskDraftChecklistItem[];
};
