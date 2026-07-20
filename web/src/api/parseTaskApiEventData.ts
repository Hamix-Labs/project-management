import type {
  ChecklistItemRemovedEventData,
  CycleLifecycleEventData,
  CycleLifecycleEventType,
  PhaseLifecycleEventData,
  PhaseLifecycleEventType,
  TaskEventType,
  TaskPickupFailedEventData,
  TaskRetryRequestedEventData,
  TransitionEventData,
  TransitionEventType,
} from "@/types";
import { isRecord, parseFiniteNumber, parseString } from "./parseTaskApiCore";

const PHASE_LIFECYCLE_TYPES = new Set<string>([
  "phase_started",
  "phase_completed",
  "phase_failed",
  "phase_skipped",
]);

const CYCLE_LIFECYCLE_TYPES = new Set<string>([
  "cycle_started",
  "cycle_completed",
  "cycle_failed",
]);

const TRANSITION_TYPES = new Set<string>([
  "status_changed",
  "priority_changed",
  "message_added",
  "prompt_appended",
]);

function parseOptionalString(
  value: unknown,
  field: string,
): string | undefined {
  if (value === undefined || value === null) return undefined;
  return parseString(value, field);
}

function parseOptionalFiniteNumber(
  value: unknown,
  field: string,
): number | undefined {
  if (value === undefined || value === null) return undefined;
  return parseFiniteNumber(value, field);
}

function parseOptionalDetails(
  value: unknown,
  field: string,
): Record<string, unknown> | undefined {
  if (value === undefined || value === null) return undefined;
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: ${field} must be an object`);
  }
  return value;
}

function requireEventDataObject(raw: unknown): Record<string, unknown> {
  if (raw == null) return {};
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: event data must be an object");
  }
  return raw;
}

export function parsePhaseLifecycleEventData(
  raw: unknown,
): PhaseLifecycleEventData {
  const o = requireEventDataObject(raw);
  const out: PhaseLifecycleEventData = {};
  const phase = parseOptionalString(o.phase, "data.phase");
  if (phase !== undefined) out.phase = phase;
  const status = parseOptionalString(o.status, "data.status");
  if (status !== undefined) out.status = status;
  const summary = parseOptionalString(o.summary, "data.summary");
  if (summary !== undefined) out.summary = summary;
  const cycleId = parseOptionalString(o.cycle_id, "data.cycle_id");
  if (cycleId !== undefined) out.cycle_id = cycleId;
  const phaseSeq = parseOptionalFiniteNumber(o.phase_seq, "data.phase_seq");
  if (phaseSeq !== undefined) out.phase_seq = phaseSeq;
  const details = parseOptionalDetails(o.details, "data.details");
  if (details !== undefined) out.details = details;
  return out;
}

export function parseCycleLifecycleEventData(
  raw: unknown,
): CycleLifecycleEventData {
  const o = requireEventDataObject(raw);
  const out: CycleLifecycleEventData = {};
  const cycleId = parseOptionalString(o.cycle_id, "data.cycle_id");
  if (cycleId !== undefined) out.cycle_id = cycleId;
  const attemptSeq = parseOptionalFiniteNumber(
    o.attempt_seq,
    "data.attempt_seq",
  );
  if (attemptSeq !== undefined) out.attempt_seq = attemptSeq;
  const status = parseOptionalString(o.status, "data.status");
  if (status !== undefined) out.status = status;
  const reason = parseOptionalString(o.reason, "data.reason");
  if (reason !== undefined) out.reason = reason;
  const failureSummary = parseOptionalString(
    o.failure_summary,
    "data.failure_summary",
  );
  if (failureSummary !== undefined) out.failure_summary = failureSummary;
  return out;
}

export function parseTransitionEventData(raw: unknown): TransitionEventData {
  const o = requireEventDataObject(raw);
  const out: TransitionEventData = {};
  const from = parseOptionalString(o.from, "data.from");
  if (from !== undefined) out.from = from;
  const to = parseOptionalString(o.to, "data.to");
  if (to !== undefined) out.to = to;
  return out;
}

export function parseChecklistItemRemovedEventData(
  raw: unknown,
): ChecklistItemRemovedEventData {
  const o = requireEventDataObject(raw);
  const out: ChecklistItemRemovedEventData = {};
  const text = parseOptionalString(o.text, "data.text");
  if (text !== undefined) out.text = text;
  return out;
}

export function parseTaskRetryRequestedEventData(
  raw: unknown,
): TaskRetryRequestedEventData {
  const o = requireEventDataObject(raw);
  const out: TaskRetryRequestedEventData = {};
  const mode = parseOptionalString(o.mode, "data.mode");
  if (mode !== undefined) out.mode = mode;
  return out;
}

export function parseTaskPickupFailedEventData(
  raw: unknown,
): TaskPickupFailedEventData {
  const o = requireEventDataObject(raw);
  const out: TaskPickupFailedEventData = {};
  const reason = parseOptionalString(o.reason, "data.reason");
  if (reason !== undefined) out.reason = reason;
  return out;
}

/** Untyped long-tail payloads: object only; field shapes left to UI. */
export function parseUntypedEventData(raw: unknown): Record<string, unknown> {
  return requireEventDataObject(raw);
}

/**
 * Family/per-type parsers for high-traffic event `data` bags.
 * Throws when a present field has the wrong JSON type.
 */
export function parseTaskEventData(
  type: TaskEventType,
  raw: unknown,
):
  | PhaseLifecycleEventData
  | CycleLifecycleEventData
  | TransitionEventData
  | ChecklistItemRemovedEventData
  | TaskRetryRequestedEventData
  | TaskPickupFailedEventData
  | Record<string, unknown> {
  if (PHASE_LIFECYCLE_TYPES.has(type)) {
    return parsePhaseLifecycleEventData(raw);
  }
  if (CYCLE_LIFECYCLE_TYPES.has(type)) {
    return parseCycleLifecycleEventData(raw);
  }
  if (TRANSITION_TYPES.has(type)) {
    return parseTransitionEventData(raw);
  }
  if (type === "checklist_item_removed") {
    return parseChecklistItemRemovedEventData(raw);
  }
  if (type === "task_retry_requested") {
    return parseTaskRetryRequestedEventData(raw);
  }
  if (type === "task_pickup_failed") {
    return parseTaskPickupFailedEventData(raw);
  }
  return parseUntypedEventData(raw);
}

export function isPhaseLifecycleEventType(
  type: TaskEventType,
): type is PhaseLifecycleEventType {
  return PHASE_LIFECYCLE_TYPES.has(type);
}

export function isCycleLifecycleEventType(
  type: TaskEventType,
): type is CycleLifecycleEventType {
  return CYCLE_LIFECYCLE_TYPES.has(type);
}

export function isTransitionEventType(
  type: TaskEventType,
): type is TransitionEventType {
  return TRANSITION_TYPES.has(type);
}
