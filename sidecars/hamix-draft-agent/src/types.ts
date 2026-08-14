// Wire types shared between the Go domain and the sidecar. Names and shapes
// mirror pkgs/draftassist/domain so the SSE frames flow through taskapi
// unchanged.

export type RunStatus =
  | "idle"
  | "thinking"
  | "streaming"
  | "tool"
  | "cancelling"
  | "cancelled"
  | "done"
  | "failed";

export type EventKind =
  | "session"
  | "status"
  | "token"
  | "tool"
  | "patch"
  | "error"
  | "done";

export interface FormSnapshot {
  title?: string;
  prompt?: string;
  priority?: string;
  project_id?: string;
  criteria?: string[];
  tags?: string[];
  cursor_model?: string;
}

export interface RunRequestBody {
  session_id: string;
  run_id?: string;
  user_message: string;
  snapshot?: FormSnapshot;
  worktree_cwd: string;
  model?: string;
  agent_id?: string;
}

export interface StatusEventData {
  status: RunStatus;
  reason?: string;
}

export interface TokenEventData {
  delta: string;
}

export interface ToolEventData {
  name: string;
  phase: "start" | "end";
  ok?: boolean;
  error?: string;
}

export type PatchOp = "set" | "find_replace" | "append";

export interface PatchEventData {
  op: PatchOp;
  find?: string;
  value?: string;
  summary?: string;
}

export interface ErrorEventData {
  code: string;
  message: string;
}

export interface DoneEventData {
  status: RunStatus;
}

export interface SessionEventData {
  session_id: string;
  worktree_id?: string;
  snapshot: FormSnapshot;
  schema_version: number;
}

export const SESSION_SCHEMA_VERSION = 1;

export interface BindFile {
  bind_schema_version: 1;
  session_id: string;
  nonce: string;
  taskapi_base_url?: string;
}

export type EmitEvent =
  | { kind: "session"; data: SessionEventData }
  | { kind: "status"; data: StatusEventData }
  | { kind: "token"; data: TokenEventData }
  | { kind: "tool"; data: ToolEventData }
  | { kind: "patch"; data: PatchEventData }
  | { kind: "error"; data: ErrorEventData }
  | { kind: "done"; data: DoneEventData };
