import { statusListLabel } from "@/lib/taskStatusDisplay";
import { parseVerificationSnapshot } from "../../task-events/parseVerificationSnapshot";
import type { Status, TaskActivityEvent } from "@/types";
import { STATUSES } from "@/types";
import type { TimelineEvent, TimelineKind } from "./timelineTypes";

/**
 * Unique id for a mapped TimelineEvent derived from the server-side
 * task_id + seq combination (stable across re-renders).
 */
function activityEventId(ev: TaskActivityEvent): string {
  return `activity-${ev.task_id}-${ev.seq}`;
}

/** Short task reference (first 8 chars of the uuid). */
function taskRef(taskId: string): string {
  return taskId.replace(/-/g, "").slice(0, 8);
}

function isStatus(s: unknown): s is Status {
  return typeof s === "string" && (STATUSES as readonly string[]).includes(s);
}

// ---------------------------------------------------------------------------
// status_changed
// ---------------------------------------------------------------------------

function mapStatusChanged(ev: TaskActivityEvent): TimelineEvent {
  const from = isStatus(ev.data.from) ? ev.data.from : "";
  const to = isStatus(ev.data.to) ? ev.data.to : "";

  const fromLabel = from ? statusListLabel(from) : from;
  const toLabel = to ? statusListLabel(to) : to;

  const highlight = ev.task_title ?? taskRef(ev.task_id);
  const detail =
    fromLabel && toLabel
      ? `${fromLabel} → ${toLabel}`
      : "Status updated.";
  const meta: string[] = [];
  if (fromLabel && toLabel) {
    meta.push(`${fromLabel} → ${toLabel}`);
  }

  return {
    id: activityEventId(ev),
    kind: "status-changed" as TimelineKind,
    category: "tasks",
    at: ev.at,
    title: "Status changed",
    highlight,
    detail,
    taskId: ev.task_id,
    taskRef: taskRef(ev.task_id),
    seq: ev.seq,
    meta: meta.length > 0 ? meta : undefined,
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
  const summary =
    typeof ev.data.failure_summary === "string" && ev.data.failure_summary
      ? ev.data.failure_summary
      : typeof ev.data.summary === "string" && ev.data.summary
        ? ev.data.summary
        : "";

  const snapshot = parseVerificationSnapshot(ev.data.details);

  const phaseLabel = KNOWN_PHASES.has(phase)
    ? phaseDisplayLabel(phase)
    : "Phase";
  const highlight = ev.task_title ?? taskRef(ev.task_id);
  const detail =
    summary || (snapshot ? `${snapshot.failedCount} criteria failed` : `${phaseLabel} phase failed.`);

  const meta: string[] = [];
  if (snapshot) {
    meta.push(`${snapshot.passedCount} passed`, `${snapshot.failedCount} failed`);
  } else if (summary) {
    meta.push(`${phaseLabel} failed`);
  }

  return {
    id: activityEventId(ev),
    kind: "verification-failed" as TimelineKind,
    category: "verification",
    at: ev.at,
    title: `${phaseLabel} phase failed`,
    highlight,
    detail,
    taskId: ev.task_id,
    taskRef: taskRef(ev.task_id),
    seq: ev.seq,
    meta: meta.length > 0 ? meta : undefined,
  };
}

// ---------------------------------------------------------------------------
// approval_granted
// ---------------------------------------------------------------------------

function mapApprovalGranted(ev: TaskActivityEvent): TimelineEvent {
  const highlight = ev.task_title ?? taskRef(ev.task_id);
  return {
    id: activityEventId(ev),
    kind: "review-approved" as TimelineKind,
    category: "tasks",
    at: ev.at,
    title: "Review approved",
    highlight,
    detail: "Approval granted — task cleared for the next step.",
    taskId: ev.task_id,
    taskRef: taskRef(ev.task_id),
    seq: ev.seq,
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
