import type { ReactNode } from "react";
import type { DraftAssistPatchOp, DraftAssistSnapshot } from "@/types/draftAssist";
import type { DraftAssistStatusState } from "./draftAssistStatus";

/** Message rendered in the assist thread. */
export type DraftAssistThreadMessage =
  | {
      id: string;
      kind: "user";
      text: string;
      at: string;
    }
  | {
      id: string;
      kind: "assistant";
      text: string;
      at: string;
      /** True once a `done` frame closes this turn. */
      done: boolean;
    }
  | {
      id: string;
      kind: "tool";
      name: string;
      phase: "start" | "end";
      ok?: boolean;
      error?: string;
      at: string;
    }
  | {
      id: string;
      kind: "patch";
      op: DraftAssistPatchOp;
      summary?: string;
      applied: boolean;
      at: string;
    }
  | {
      id: string;
      kind: "error";
      message: string;
      at: string;
    };

export type DraftAssistContextValue = {
  /** True once the operator has triggered the assist column at least once. */
  active: boolean;
  status: DraftAssistStatusState;
  messages: DraftAssistThreadMessage[];
  /** Session id from POST /sessions once known. */
  sessionId: string | null;
  /** Timestamp of the current run for the watchdog (null when idle). */
  runStartedAt: number | null;
  /** True while a run is in flight (SSE not yet terminal). */
  runActive: boolean;
  /**
   * Open the assist column and (lazily) start the session, then send the
   * first user message. Subsequent triggers should call {@link send}.
   */
  open: (userMessage: string) => void;
  /** Send a follow-up on the same session. */
  send: (userMessage: string) => void;
  /** Cancel the in-flight run (idempotent when no run is active). */
  stop: () => void;
  /** Best-effort teardown; also fires on unmount via `useDraftAssistSession`. */
  reset: () => void;
};

export type DraftAssistProviderProps = {
  /** Read the compose form's current draft — used for POST /sessions and follow-ups. */
  getSnapshot: () => DraftAssistSnapshot;
  /** Optional worktree binding — forwarded to POST /sessions. */
  worktreeId?: string;
  /** Replace the prompt string when a `patch` frame arrives. */
  onApplyPromptPatch?: (nextPrompt: string) => void;
  /** Read the current prompt so `find_replace` patches can operate on the live value. */
  getPromptSnapshot?: () => string;
  children: ReactNode;
};
