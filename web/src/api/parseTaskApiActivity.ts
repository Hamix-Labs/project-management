import type {
  Priority,
  TaskActivityEvent,
  TaskActivityEventType,
  TaskActivityResponse,
} from "@/types";
import { PRIORITIES } from "@/types";

const ACTIVITY_EVENT_TYPES: readonly TaskActivityEventType[] = [
  "status_changed",
  "phase_failed",
  "approval_granted",
];

function isActivityEventType(v: unknown): v is TaskActivityEventType {
  return (
    typeof v === "string" &&
    (ACTIVITY_EVENT_TYPES as readonly string[]).includes(v)
  );
}

function parseOptionalPriority(v: unknown): Priority | undefined {
  if (typeof v !== "string") return undefined;
  if (!(PRIORITIES as readonly string[]).includes(v)) return undefined;
  return v as Priority;
}

function parseOptionalTags(v: unknown): string[] | undefined {
  if (!Array.isArray(v)) return undefined;
  const tags: string[] = [];
  for (const raw of v) {
    if (typeof raw === "string" && raw.trim()) {
      tags.push(raw.trim());
    }
  }
  return tags;
}

function parseActivityEvent(raw: unknown): TaskActivityEvent | null {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) return null;
  const o = raw as Record<string, unknown>;

  const task_id = typeof o.task_id === "string" ? o.task_id.trim() : "";
  if (!task_id) return null;

  const seq =
    typeof o.seq === "number" && Number.isFinite(o.seq) && o.seq >= 1
      ? o.seq
      : null;
  if (seq === null) return null;

  const at = typeof o.at === "string" ? o.at.trim() : "";
  if (!at) return null;

  if (!isActivityEventType(o.type)) return null;

  const by = typeof o.by === "string" ? o.by.trim() : "";

  let data: Record<string, unknown> = {};
  if (o.data && typeof o.data === "object" && !Array.isArray(o.data)) {
    data = o.data as Record<string, unknown>;
  }

  const task_title =
    typeof o.task_title === "string" && o.task_title.trim()
      ? o.task_title.trim()
      : undefined;

  const task_number =
    typeof o.task_number === "number" && Number.isFinite(o.task_number)
      ? o.task_number
      : undefined;

  const task_priority = parseOptionalPriority(o.task_priority);

  const task_project_id =
    typeof o.task_project_id === "string" && o.task_project_id.trim()
      ? o.task_project_id.trim()
      : undefined;

  const task_tags = parseOptionalTags(o.task_tags);

  return {
    task_id,
    seq,
    at,
    type: o.type,
    by,
    data,
    task_title,
    task_number,
    task_priority,
    task_project_id,
    task_tags,
  };
}

export function parseTaskActivityResponse(raw: unknown): TaskActivityResponse {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    throw new Error("parseTaskActivityResponse: expected object");
  }
  const o = raw as Record<string, unknown>;

  const total =
    typeof o.total === "number" && Number.isFinite(o.total) && o.total >= 0
      ? o.total
      : 0;
  const limit =
    typeof o.limit === "number" && Number.isFinite(o.limit) && o.limit >= 1
      ? o.limit
      : 50;
  const offset =
    typeof o.offset === "number" &&
    Number.isFinite(o.offset) &&
    o.offset >= 0
      ? o.offset
      : 0;

  const rawEvents = Array.isArray(o.events) ? o.events : [];
  const events: TaskActivityEvent[] = [];
  for (const row of rawEvents) {
    const ev = parseActivityEvent(row);
    if (ev) events.push(ev);
  }

  return { total, limit, offset, events };
}
