import { statusListLabel } from "@/lib/taskStatusDisplay";
import { taskDisplayRef } from "@/lib/taskShortId";
import type { Priority, Status, TaskActivityEvent } from "@/types";
import { PRIORITIES, STATUSES } from "@/types";
import { parseVerificationSnapshot } from "../../task-events/parseVerificationSnapshot";
import {
  firstFailedCriterionHighlight,
  phaseFailureDetail,
} from "./phaseFailureDetail";
import type { TimelineEvent, TimelineKind } from "./timelineTypes";

/**
 * Unique id for a mapped TimelineEvent derived from the server-side
 * task_id + seq combination (stable across re-renders).
 */
function activityEventId(ev: TaskActivityEvent): string {
  return `activity-${ev.task_id}-${ev.seq}`;
}

/**
 * Canonical short reference for a timeline row: prefers the per-project
 * `#N` when the server surfaced `task_number`, else the short UUID.
 * `task_title` is intentionally NOT used in the title/highlight — see
 * the Timeline spec: titles blow up the layout / leak edited titles
 * into the audit trail. Title is kept on the envelope for search only.
 */
function taskRef(ev: Pick<TaskActivityEvent, "task_id" | "task_number">): string {
  return taskDisplayRef({ id: ev.task_id, number: ev.task_number });
}

function isStatus(s: unknown): s is Status {
  return typeof s === "string" && (STATUSES as readonly string[]).includes(s);
}

function isPriority(v: unknown): v is Priority {
  return typeof v === "string" && (PRIORITIES as readonly string[]).includes(v);
}

function filterFields(ev: TaskActivityEvent): Pick<
  TimelineEvent,
  "taskTitle" | "taskPriority" | "taskProjectId" | "taskTags"
> {
  return {
    taskTitle: ev.task_title,
    taskPriority: isPriority(ev.task_priority) ? ev.task_priority : undefined,
    taskProjectId: ev.task_project_id ?? undefined,
    taskTags: ev.task_tags,
  };
}

// ---------------------------------------------------------------------------
// status_changed
// ---------------------------------------------------------------------------

function mapStatusChanged(ev: TaskActivityEvent): TimelineEvent {
  const from = isStatus(ev.data.from) ? ev.data.from : "";
  const to = isStatus(ev.data.to) ? ev.data.to : "";

  const fromLabel = from ? statusListLabel(from) : from;
  const toLabel = to ? statusListLabel(to) : to;

  const ref = taskRef(ev);
  const detail =
    fromLabel && toLabel
      ? `${fromLabel} → ${toLabel}`
      : "Status updated.";

  return {
    id: activityEventId(ev),
    kind: "status-changed" as TimelineKind,
    category: "tasks",
    at: ev.at,
    title: "Status changed",
    highlight: "",
    detail,
    taskId: ev.task_id,
    taskRef: ref,
    seq: ev.seq,
    ...filterFields(ev),
  };
}

// ---------------------------------------------------------------------------
// phase_failed
// ---------------------------------------------------------------------------

const KNOWN_PHASES = new Set(["execute", "verify", "diagnose", "persist"]);

function phaseDisplayLabel(phase: unknown): string {
  if (typeof phase !== "string") return "Phase";
  switch (phase) {
    case "execute":
      return "Execute";
    case "verify":
      return "Verify";
    case "diagnose":
      return "Diagnose";
    case "persist":
      return "Persist";
    default:
      return "Phase";
  }
}

function mapPhaseFailed(ev: TaskActivityEvent): TimelineEvent {
  const phase = typeof ev.data.phase === "string" ? ev.data.phase : "";
  const snapshot = parseVerificationSnapshot(ev.data.details);

  const phaseLabel = KNOWN_PHASES.has(phase)
    ? phaseDisplayLabel(phase)
    : "Phase";
  const ref = taskRef(ev);
  const detail = phaseFailureDetail(ev.data, phaseLabel, snapshot);
  const highlight = firstFailedCriterionHighlight(snapshot);

  const meta: string[] = [];
  if (snapshot) {
    meta.push(`${snapshot.passedCount} passed`, `${snapshot.failedCount} failed`);
  }

  const title =
    phase === "verify" ? "Verification failed" : `${phaseLabel} phase failed`;

  return {
    id: activityEventId(ev),
    kind: "verification-failed" as TimelineKind,
    category: "verification",
    at: ev.at,
    title,
    highlight,
    detail,
    taskId: ev.task_id,
    taskRef: ref,
    seq: ev.seq,
    meta: meta.length > 0 ? meta : undefined,
    ...filterFields(ev),
  };
}

// ---------------------------------------------------------------------------
// approval_granted
// ---------------------------------------------------------------------------

function mapApprovalGranted(ev: TaskActivityEvent): TimelineEvent {
  const ref = taskRef(ev);
  return {
    id: activityEventId(ev),
    kind: "review-approved" as TimelineKind,
    category: "tasks",
    at: ev.at,
    title: "Review approved",
    highlight: "",
    detail: "Approval granted — task cleared for the next step.",
    taskId: ev.task_id,
    taskRef: ref,
    seq: ev.seq,
    ...filterFields(ev),
  };
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

/**
 * Maps a raw ActivityEvent from `GET /tasks/activity` to a `TimelineEvent`
 * suitable for the home Timeline view. Returns `null` for unknown types.
 */
export function mapActivityEventToTimeline(
  ev: TaskActivityEvent,
): TimelineEvent | null {
  switch (ev.type) {
    case "status_changed":
      return mapStatusChanged(ev);
    case "phase_failed":
      return mapPhaseFailed(ev);
    case "approval_granted":
      return mapApprovalGranted(ev);
    default:
      return null;
  }
}

/**
 * Maps an array of raw ActivityEvents to TimelineEvents, dropping nulls.
 */
export function mapActivityEventsToTimeline(
  events: TaskActivityEvent[],
): TimelineEvent[] {
  const out: TimelineEvent[] = [];
  for (const ev of events) {
    const mapped = mapActivityEventToTimeline(ev);
    if (mapped) out.push(mapped);
  }
  return out;
}
