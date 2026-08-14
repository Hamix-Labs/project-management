import type {
  DraftAssistEvent,
  DraftAssistRunStatus,
} from "@/types/draftAssist";

/**
 * UI-facing status for the assist thread. Mirrors the client status
 * machine documented in `docs/design/task-draft-ai.md` §Status machine:
 *
 *   idle → starting → thinking → streaming | tool → applying → idle
 *
 * Plus the cross-cutting states: `cancelling`, `error`, `disconnected`.
 * Copy strings must match the design (sentence case, specific) so the
 * a11y live region never reads a generic "Loading…".
 */
export type DraftAssistUiStatus =
  | "idle"
  | "starting"
  | "thinking"
  | "streaming"
  | "tool"
  | "applying"
  | "cancelling"
  | "error"
  | "disconnected";

/** Terminal message rendered once a run finishes (falls through to `idle`). */
export type DraftAssistTerminalNote =
  | "prompt_updated"
  | "assistant_stopped"
  | null;

/** Reducer state — kept minimal; view layer derives copy via {@link draftAssistStatusCopy}. */
export type DraftAssistStatusState = {
  status: DraftAssistUiStatus;
  /** Set by `tool` frames so the copy can say `Reading path…`. */
  toolName: string | null;
  /** Optional short suffix (e.g. patch summary). */
  detail: string | null;
  /** Last error message from `error` frame or `done{failed}`. */
  errorMessage: string | null;
  /** Terminal note carried forward after `done` so the thread can label it. */
  terminal: DraftAssistTerminalNote;
};

export const INITIAL_DRAFT_ASSIST_STATUS: DraftAssistStatusState = {
  status: "idle",
  toolName: null,
  detail: null,
  errorMessage: null,
  terminal: null,
};

export type DraftAssistStatusAction =
  /** Composer Send fired — POST /runs is in flight. */
  | { type: "run_requested" }
  /** POST /runs resolved 202; we now wait for the first SSE frame. */
  | { type: "run_accepted" }
  /** Composer Stop fired — POST /cancel is in flight. */
  | { type: "cancel_requested" }
  /** Any SSE frame from `useDraftAssistStream`. */
  | { type: "event"; event: DraftAssistEvent }
  /** Stream transport transitioned (connecting/live/disconnected). */
  | { type: "connection"; connected: boolean }
  /** POST /runs or POST /cancel rejected before any SSE arrived. */
  | { type: "transport_error"; message: string }
  /** Compose page unmount or explicit reset — back to idle. */
  | { type: "reset" };

/**
 * Reducer: SSE event or transport action → next UI status.
 *
 * Design constraint: never regress from `applying`/`streaming` into
 * `thinking` for `tool end` — hold whatever the previous "worked-on"
 * status was so the thread never reads like it started over.
 */
export function draftAssistStatusReducer(
  state: DraftAssistStatusState,
  action: DraftAssistStatusAction,
): DraftAssistStatusState {
  switch (action.type) {
    case "reset":
      return INITIAL_DRAFT_ASSIST_STATUS;

    case "run_requested":
      return {
        ...INITIAL_DRAFT_ASSIST_STATUS,
        status: "starting",
      };

    case "run_accepted":
      // Held at `starting` until the first `status` or `token` frame; if
      // the run was already advanced by earlier frames leave it alone.
      if (state.status === "idle") return { ...state, status: "starting" };
      return state;

    case "cancel_requested":
      return { ...state, status: "cancelling", errorMessage: null };

    case "connection":
      if (action.connected) {
        // Recovered from a disconnect — re-enter a working state; the next
        // SSE frame will refine it. If we were never mid-run, stay idle.
        if (state.status === "disconnected") {
          return { ...state, status: "thinking" };
        }
        return state;
      }
      // Only degrade to `disconnected` when a run is in flight; a dropped
      // idle socket does not need to shout at the operator.
      if (state.status === "idle") return state;
      return { ...state, status: "disconnected" };

    case "transport_error":
      return {
        ...state,
        status: "error",
        errorMessage: action.message,
      };

    case "event":
      return reduceEvent(state, action.event);

    default:
      return state;
  }
}

function reduceEvent(
  state: DraftAssistStatusState,
  event: DraftAssistEvent,
): DraftAssistStatusState {
  switch (event.kind) {
    case "session":
      // Handshake — do not move status. Callers use this to record the
      // session id and validate schema.
      return state;

    case "status": {
      const next = uiStatusFromRunStatus(event.data.status);
      if (next === null) return state;
      if (next === "cancelling") {
        return { ...state, status: "cancelling" };
      }
      return { ...state, status: next, toolName: null, errorMessage: null };
    }

    case "token":
      // Any token means streaming; clear tool overlay so copy no longer
      // reads `Reading foo.ts…` after the assistant starts writing.
      return { ...state, status: "streaming", toolName: null };

    case "tool": {
      if (event.data.phase === "start") {
        return { ...state, status: "tool", toolName: event.data.name };
      }
      // Tool ended: hold the previous "worked-on" state until a fresh
      // status/token frame arrives; drop the tool name so copy stops
      // saying `Reading …`.
      const held: DraftAssistUiStatus =
        state.status === "tool" ? "thinking" : state.status;
      return { ...state, status: held, toolName: null };
    }

    case "patch":
      return {
        ...state,
        status: "applying",
        detail: event.data.summary ?? null,
      };

    case "error":
      return {
        ...state,
        status: "error",
        errorMessage: event.data.message,
      };

    case "done": {
      if (event.data.status === "failed") {
        return {
          ...state,
          status: "error",
          errorMessage: state.errorMessage ?? "The assistant reported an error.",
          terminal: null,
        };
      }
      if (event.data.status === "cancelled") {
        return {
          ...INITIAL_DRAFT_ASSIST_STATUS,
          terminal: "assistant_stopped",
        };
      }
      // done{done} — happy terminal. Keep any patch-applied detail so the
      // thread can end on `Prompt updated`.
      return {
        ...INITIAL_DRAFT_ASSIST_STATUS,
        terminal: state.detail ? "prompt_updated" : null,
      };
    }
  }
}

function uiStatusFromRunStatus(
  s: DraftAssistRunStatus,
): DraftAssistUiStatus | null {
  switch (s) {
    case "idle":
      return "idle";
    case "thinking":
      return "thinking";
    case "streaming":
      return "streaming";
    case "tool":
      return "tool";
    case "cancelling":
      return "cancelling";
    case "cancelled":
    case "done":
    case "failed":
      // Terminal statuses are handled through `done` frames.
      return null;
  }
}

/**
 * Human copy for `aria-live` region and the status pill. Sentence case,
 * specific — matches the design copy set (`Starting assistant…`,
 * `Thinking…`, `Reading path…`, `Updating prompt…`, `Assistant stopped`,
 * `Prompt updated`, `Couldn’t reach the assistant. Retry`).
 */
export function draftAssistStatusCopy(state: DraftAssistStatusState): string {
  switch (state.status) {
    case "idle":
      if (state.terminal === "prompt_updated") return "Prompt updated";
      if (state.terminal === "assistant_stopped") return "Assistant stopped";
      return "";
    case "starting":
      return "Starting assistant…";
    case "thinking":
      return "Thinking…";
    case "streaming":
      return "Writing…";
    case "tool":
      if (state.toolName) return `Reading ${state.toolName}…`;
      return "Working…";
    case "applying":
      return state.detail ? `Updating prompt — ${state.detail}` : "Updating prompt…";
    case "cancelling":
      return "Stopping…";
    case "disconnected":
      return "Reconnecting…";
    case "error":
      return state.errorMessage
        ? `Couldn’t reach the assistant. ${state.errorMessage}`
        : "Couldn’t reach the assistant. Retry.";
  }
}

/**
 * True while the assistant is actively running — used to swap the
 * Send button for Stop, drive the watchdog, and gate cancellation.
 */
export function isDraftAssistRunActive(state: DraftAssistStatusState): boolean {
  switch (state.status) {
    case "starting":
    case "thinking":
    case "streaming":
    case "tool":
    case "applying":
    case "cancelling":
    case "disconnected":
      return true;
    case "idle":
    case "error":
      return false;
  }
}
