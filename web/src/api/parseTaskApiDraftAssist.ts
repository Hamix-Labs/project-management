import {
  DRAFT_ASSIST_EVENT_KINDS,
  DRAFT_ASSIST_NOT_READY_REASONS,
  DRAFT_ASSIST_PATCH_OPS,
  DRAFT_ASSIST_RUNNER_NAMES,
  DRAFT_ASSIST_RUN_STATUSES,
  DRAFT_ASSIST_SCHEMA_VERSION,
  type DraftAssistCancelRunResult,
  type DraftAssistDoneEventData,
  type DraftAssistErrorEventData,
  type DraftAssistEvent,
  type DraftAssistEventKind,
  type DraftAssistNotReadyReason,
  type DraftAssistPatchEventData,
  type DraftAssistPatchOp,
  type DraftAssistReady,
  type DraftAssistRunStatus,
  type DraftAssistRunnerName,
  type DraftAssistSession,
  type DraftAssistSessionEventData,
  type DraftAssistSnapshot,
  type DraftAssistSnapshotUpdate,
  type DraftAssistStartRunResult,
  type DraftAssistStatusEventData,
  type DraftAssistTokenEventData,
  type DraftAssistToolEventData,
} from "@/types/draftAssist";
import {
  isRecord,
  parseBooleanField,
  parseFiniteNumber,
  parseISO8601Required,
  parseNonEmptyString,
  parseOptionalNonEmptyId,
  parseOptionalParseableDate,
  parseString,
} from "./parseTaskApiCore";

function parseOptionalString(v: unknown, field: string): string | undefined {
  if (v === undefined || v === null || v === "") return undefined;
  return parseString(v, field);
}

function parseStringArray(v: unknown, field: string): string[] | undefined {
  if (v === undefined || v === null) return undefined;
  if (!Array.isArray(v)) {
    throw new Error(`Invalid API response: ${field} must be an array of strings`);
  }
  return v.map((item, i) => parseString(item, `${field}[${i}]`));
}

/** Validates a wire snapshot object; every field is optional per Go domain. */
export function parseDraftAssistSnapshot(value: unknown): DraftAssistSnapshot {
  if (value === undefined || value === null) return {};
  if (!isRecord(value)) {
    throw new Error("Invalid API response: snapshot must be an object");
  }
  const out: DraftAssistSnapshot = {};
  const title = parseOptionalString(value.title, "snapshot.title");
  if (title !== undefined) out.title = title;
  const prompt = parseOptionalString(value.prompt, "snapshot.prompt");
  if (prompt !== undefined) out.prompt = prompt;
  const priority = parseOptionalString(value.priority, "snapshot.priority");
  if (priority !== undefined) out.priority = priority;
  const projectId = parseOptionalString(value.project_id, "snapshot.project_id");
  if (projectId !== undefined) out.project_id = projectId;
  const criteria = parseStringArray(value.criteria, "snapshot.criteria");
  if (criteria !== undefined) out.criteria = criteria;
  const tags = parseStringArray(value.tags, "snapshot.tags");
  if (tags !== undefined) out.tags = tags;
  const cursorModel = parseOptionalString(value.cursor_model, "snapshot.cursor_model");
  if (cursorModel !== undefined) out.cursor_model = cursorModel;
  const updatedAt = parseOptionalParseableDate(value.updated_at, "snapshot.updated_at");
  if (updatedAt !== undefined) out.updated_at = updatedAt;
  return out;
}

/** Validates POST /draft-assist/sessions and GET /draft-assist/sessions/{id} JSON. */
export function parseDraftAssistSession(value: unknown): DraftAssistSession {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: session must be an object");
  }
  const out: DraftAssistSession = {
    id: parseNonEmptyString(value.id, "id"),
    nonce: parseNonEmptyString(value.nonce, "nonce"),
    snapshot: parseDraftAssistSnapshot(value.snapshot),
    created_at: parseISO8601Required(value.created_at, "created_at"),
  };
  const worktreeId = parseOptionalNonEmptyId(value.worktree_id, "worktree_id");
  if (worktreeId !== undefined) out.worktree_id = worktreeId;
  const updatedAt = parseOptionalParseableDate(value.updated_at, "updated_at");
  if (updatedAt !== undefined) out.updated_at = updatedAt;
  return out;
}

/** Validates PUT /draft-assist/sessions/{id}/snapshot JSON. */
export function parseDraftAssistSnapshotUpdate(value: unknown): DraftAssistSnapshotUpdate {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: snapshot update must be an object");
  }
  return {
    id: parseNonEmptyString(value.id, "id"),
    snapshot: parseDraftAssistSnapshot(value.snapshot),
  };
}

/** Validates POST /draft-assist/sessions/{id}/runs JSON. */
export function parseDraftAssistStartRunResult(value: unknown): DraftAssistStartRunResult {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: start-run response must be an object");
  }
  return { run_id: parseNonEmptyString(value.run_id, "run_id") };
}

/** Validates POST /draft-assist/sessions/{id}/runs/{runId}/cancel JSON. */
export function parseDraftAssistCancelRunResult(value: unknown): DraftAssistCancelRunResult {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: cancel-run response must be an object");
  }
  const status = parseString(value.status, "status");
  if (status !== "cancelling") {
    throw new Error("Invalid API response: cancel status must be 'cancelling'");
  }
  return {
    run_id: parseNonEmptyString(value.run_id, "run_id"),
    status: "cancelling",
  };
}

function parseRunnerName(value: unknown): DraftAssistRunnerName {
  const s = parseString(value, "runner");
  if (!(DRAFT_ASSIST_RUNNER_NAMES as readonly string[]).includes(s)) {
    throw new Error(`Invalid API response: runner must be one of ${DRAFT_ASSIST_RUNNER_NAMES.join("|")}`);
  }
  return s as DraftAssistRunnerName;
}

function parseNotReadyReason(value: unknown): DraftAssistNotReadyReason {
  const s = parseString(value, "reason");
  if (!(DRAFT_ASSIST_NOT_READY_REASONS as readonly string[]).includes(s)) {
    throw new Error(`Invalid API response: reason must be one of ${DRAFT_ASSIST_NOT_READY_REASONS.join("|")}`);
  }
  return s as DraftAssistNotReadyReason;
}

/** Validates GET /draft-assist/ready JSON. */
export function parseDraftAssistReady(value: unknown): DraftAssistReady {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: ready must be an object");
  }
  const out: DraftAssistReady = {
    ready: parseBooleanField(value.ready, "ready"),
    runner: parseRunnerName(value.runner),
  };
  if (value.reason !== undefined && value.reason !== null && value.reason !== "") {
    out.reason = parseNotReadyReason(value.reason);
  }
  return out;
}

function parseRunStatus(value: unknown, field: string): DraftAssistRunStatus {
  const s = parseString(value, field);
  if (!(DRAFT_ASSIST_RUN_STATUSES as readonly string[]).includes(s)) {
    throw new Error(`Invalid API response: ${field} must be a known run status`);
  }
  return s as DraftAssistRunStatus;
}

function parsePatchOp(value: unknown): DraftAssistPatchOp {
  const s = parseString(value, "op");
  if (!(DRAFT_ASSIST_PATCH_OPS as readonly string[]).includes(s)) {
    throw new Error(`Invalid API response: op must be one of ${DRAFT_ASSIST_PATCH_OPS.join("|")}`);
  }
  return s as DraftAssistPatchOp;
}

function parseEventKind(value: unknown): DraftAssistEventKind {
  const s = parseString(value, "kind");
  if (!(DRAFT_ASSIST_EVENT_KINDS as readonly string[]).includes(s)) {
    throw new Error(`Invalid API response: kind must be one of ${DRAFT_ASSIST_EVENT_KINDS.join("|")}`);
  }
  return s as DraftAssistEventKind;
}

/**
 * Validates a session-event payload. Hard-throws when `schema_version`
 * does not match {@link DRAFT_ASSIST_SCHEMA_VERSION} so a mismatched
 * server never streams unsafe events into the SPA.
 */
export function parseDraftAssistSessionEventData(value: unknown): DraftAssistSessionEventData {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: session event data must be an object");
  }
  const schemaVersion = parseFiniteNumber(value.schema_version, "schema_version");
  if (schemaVersion !== DRAFT_ASSIST_SCHEMA_VERSION) {
    throw new Error(
      `Draft-assist schema mismatch: SPA expects ${DRAFT_ASSIST_SCHEMA_VERSION}, server sent ${schemaVersion}`,
    );
  }
  const out: DraftAssistSessionEventData = {
    session_id: parseNonEmptyString(value.session_id, "session_id"),
    snapshot: parseDraftAssistSnapshot(value.snapshot),
    schema_version: schemaVersion,
  };
  const worktreeId = parseOptionalNonEmptyId(value.worktree_id, "worktree_id");
  if (worktreeId !== undefined) out.worktree_id = worktreeId;
  return out;
}

function parseStatusEventData(value: unknown): DraftAssistStatusEventData {
  if (!isRecord(value)) throw new Error("Invalid API response: status data must be an object");
  const out: DraftAssistStatusEventData = { status: parseRunStatus(value.status, "status") };
  const reason = parseOptionalString(value.reason, "reason");
  if (reason !== undefined) out.reason = reason;
  return out;
}

function parseTokenEventData(value: unknown): DraftAssistTokenEventData {
  if (!isRecord(value)) throw new Error("Invalid API response: token data must be an object");
  return { delta: parseString(value.delta, "delta") };
}

function parseToolEventData(value: unknown): DraftAssistToolEventData {
  if (!isRecord(value)) throw new Error("Invalid API response: tool data must be an object");
  const phaseRaw = parseString(value.phase, "phase");
  if (phaseRaw !== "start" && phaseRaw !== "end") {
    throw new Error("Invalid API response: tool.phase must be 'start' or 'end'");
  }
  const out: DraftAssistToolEventData = {
    name: parseNonEmptyString(value.name, "name"),
    phase: phaseRaw,
  };
  if (value.ok !== undefined && value.ok !== null) {
    out.ok = parseBooleanField(value.ok, "ok");
  }
  const err = parseOptionalString(value.error, "error");
  if (err !== undefined) out.error = err;
  return out;
}

function parsePatchEventData(value: unknown): DraftAssistPatchEventData {
  if (!isRecord(value)) throw new Error("Invalid API response: patch data must be an object");
  const out: DraftAssistPatchEventData = { op: parsePatchOp(value.op) };
  const find = parseOptionalString(value.find, "find");
  if (find !== undefined) out.find = find;
  const val = parseOptionalString(value.value, "value");
  if (val !== undefined) out.value = val;
  const summary = parseOptionalString(value.summary, "summary");
  if (summary !== undefined) out.summary = summary;
  return out;
}

function parseErrorEventData(value: unknown): DraftAssistErrorEventData {
  if (!isRecord(value)) throw new Error("Invalid API response: error data must be an object");
  return {
    code: parseNonEmptyString(value.code, "code"),
    message: parseString(value.message, "message"),
  };
}

function parseDoneEventData(value: unknown): DraftAssistDoneEventData {
  if (!isRecord(value)) throw new Error("Invalid API response: done data must be an object");
  return { status: parseRunStatus(value.status, "status") };
}

/**
 * Validates one SSE envelope. Kind selects the payload parser; the
 * `session` event enforces {@link DRAFT_ASSIST_SCHEMA_VERSION}.
 */
export function parseDraftAssistEvent(value: unknown): DraftAssistEvent {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: event must be an object");
  }
  const id = parseFiniteNumber(value.id, "id");
  const kind = parseEventKind(value.kind);
  const at = parseISO8601Required(value.at, "at");
  const runId = parseOptionalNonEmptyId(value.run_id, "run_id");
  const base: { id: number; run_id?: string; at: string } = { id, at };
  if (runId !== undefined) base.run_id = runId;
  const dataRaw = "data" in value ? value.data : undefined;
  switch (kind) {
    case "session":
      return { ...base, kind, data: parseDraftAssistSessionEventData(dataRaw) };
    case "status":
      return { ...base, kind, data: parseStatusEventData(dataRaw) };
    case "token":
      return { ...base, kind, data: parseTokenEventData(dataRaw) };
    case "tool":
      return { ...base, kind, data: parseToolEventData(dataRaw) };
    case "patch":
      return { ...base, kind, data: parsePatchEventData(dataRaw) };
    case "error":
      return { ...base, kind, data: parseErrorEventData(dataRaw) };
    case "done":
      return { ...base, kind, data: parseDoneEventData(dataRaw) };
  }
}
