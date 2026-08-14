/**
 * Wire types for /draft-assist/*. Mirrors pkgs/draftassist/domain.
 * The SPA asserts `schema_version === DRAFT_ASSIST_SCHEMA_VERSION`
 * on the first SSE `session` frame before trusting the stream
 * (see docs/domain/draft-assist.md).
 */

/** Bumped whenever the wire contract changes; must match pkgs/draftassist/domain.DraftAssistSchemaVersion. */
export const DRAFT_ASSIST_SCHEMA_VERSION = 1;

/** Create-task form state the operator has entered so far. Prompt is the only MCP-writable field in v1. */
export type DraftAssistSnapshot = {
  title?: string;
  prompt?: string;
  priority?: string;
  project_id?: string;
  criteria?: string[];
  tags?: string[];
  cursor_model?: string;
  updated_at?: string;
};

/** In-memory record for one draft-assist modal returned by POST/GET /sessions. */
export type DraftAssistSession = {
  id: string;
  nonce: string;
  worktree_id?: string;
  snapshot: DraftAssistSnapshot;
  created_at: string;
  updated_at?: string;
};

/** Run status lifecycle: idle → thinking → streaming|tool → idle; cancelling → done{cancelled|failed}. */
export const DRAFT_ASSIST_RUN_STATUSES = [
  "idle",
  "thinking",
  "streaming",
  "tool",
  "cancelling",
  "cancelled",
  "done",
  "failed",
] as const;
export type DraftAssistRunStatus = (typeof DRAFT_ASSIST_RUN_STATUSES)[number];

/** Terminal run states: no further events for the run. */
export const DRAFT_ASSIST_TERMINAL_RUN_STATUSES = ["cancelled", "done", "failed"] as const satisfies readonly DraftAssistRunStatus[];

export function isTerminalDraftAssistRunStatus(status: DraftAssistRunStatus): boolean {
  return (DRAFT_ASSIST_TERMINAL_RUN_STATUSES as readonly string[]).includes(status);
}

/** SSE named-event keys emitted on the wire. Heartbeats are `: heartbeat` comments, not named events. */
export const DRAFT_ASSIST_EVENT_KINDS = [
  "session",
  "status",
  "token",
  "tool",
  "patch",
  "error",
  "done",
] as const;
export type DraftAssistEventKind = (typeof DRAFT_ASSIST_EVENT_KINDS)[number];

/** MCP prompt-mutation operations. */
export const DRAFT_ASSIST_PATCH_OPS = ["set", "find_replace", "append"] as const;
export type DraftAssistPatchOp = (typeof DRAFT_ASSIST_PATCH_OPS)[number];

export type DraftAssistSessionEventData = {
  session_id: string;
  worktree_id?: string;
  snapshot: DraftAssistSnapshot;
  schema_version: number;
};

export type DraftAssistStatusEventData = {
  status: DraftAssistRunStatus;
  reason?: string;
};

export type DraftAssistTokenEventData = {
  delta: string;
};

export type DraftAssistToolEventData = {
  name: string;
  phase: "start" | "end";
  ok?: boolean;
  error?: string;
};

export type DraftAssistPatchEventData = {
  op: DraftAssistPatchOp;
  find?: string;
  value?: string;
  summary?: string;
};

export type DraftAssistErrorEventData = {
  code: string;
  message: string;
};

export type DraftAssistDoneEventData = {
  status: DraftAssistRunStatus;
};

/** Envelope base fields present on every frame. */
type DraftAssistEventBase = {
  id: number;
  run_id?: string;
  at: string;
};

/** Tagged-union of wire events. `kind` selects the payload shape. */
export type DraftAssistEvent =
  | (DraftAssistEventBase & { kind: "session"; data: DraftAssistSessionEventData })
  | (DraftAssistEventBase & { kind: "status"; data: DraftAssistStatusEventData })
  | (DraftAssistEventBase & { kind: "token"; data: DraftAssistTokenEventData })
  | (DraftAssistEventBase & { kind: "tool"; data: DraftAssistToolEventData })
  | (DraftAssistEventBase & { kind: "patch"; data: DraftAssistPatchEventData })
  | (DraftAssistEventBase & { kind: "error"; data: DraftAssistErrorEventData })
  | (DraftAssistEventBase & { kind: "done"; data: DraftAssistDoneEventData });

/** Runner names the ready probe may return. */
export const DRAFT_ASSIST_RUNNER_NAMES = ["sdk", "fake", "missing"] as const;
export type DraftAssistRunnerName = (typeof DRAFT_ASSIST_RUNNER_NAMES)[number];

/** Reasons the ready probe may return when `ready === false`. */
export const DRAFT_ASSIST_NOT_READY_REASONS = [
  "no_runner",
  "missing_key",
  "sidecar_down",
] as const;
export type DraftAssistNotReadyReason = (typeof DRAFT_ASSIST_NOT_READY_REASONS)[number];

/** Response from GET /draft-assist/ready. */
export type DraftAssistReady = {
  ready: boolean;
  runner: DraftAssistRunnerName;
  reason?: DraftAssistNotReadyReason;
};

/** Response from POST /draft-assist/sessions/{id}/runs. */
export type DraftAssistStartRunResult = {
  run_id: string;
};

/** Response from POST /draft-assist/sessions/{id}/runs/{runId}/cancel. */
export type DraftAssistCancelRunResult = {
  run_id: string;
  status: "cancelling";
};

/** Response from PUT /draft-assist/sessions/{id}/snapshot. */
export type DraftAssistSnapshotUpdate = {
  id: string;
  snapshot: DraftAssistSnapshot;
};
