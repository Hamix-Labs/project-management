import type {
  TaskActivityEvent,
  TaskActivityEventType,
  TaskActivityResponse,
} from "@/types";

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

  return {
    task_id,
    seq,
    at,
    type: o.type,
    by,
    data,
    task_title,
    task_number,
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
