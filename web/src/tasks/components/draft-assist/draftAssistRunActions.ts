import {
  cancelRun as apiCancelRun,
  startRun as apiStartRun,
} from "@/api/draftAssist";
import type { DraftAssistSnapshot } from "@/types/draftAssist";
import type { Dispatch, MutableRefObject, SetStateAction } from "react";
import type { DraftAssistStatusAction } from "./draftAssistStatus";
import type { DraftAssistThreadMessage } from "./draftAssistTypes";

export type DraftAssistRunDeps = {
  session: {
    open: (
      snapshot: DraftAssistSnapshot,
      worktreeId?: string,
    ) => Promise<{ id: string }>;
    close: () => void;
    session: { id: string } | null;
  };
  worktreeId?: string;
  getSnapshotRef: MutableRefObject<() => DraftAssistSnapshot>;
  runIdRef: MutableRefObject<string | null>;
  assistantTurnIdRef: MutableRefObject<string | null>;
  startAbortRef: MutableRefObject<AbortController | null>;
  cancelAbortRef: MutableRefObject<AbortController | null>;
  dispatch: Dispatch<DraftAssistStatusAction>;
  setActive: Dispatch<SetStateAction<boolean>>;
  setMessages: Dispatch<SetStateAction<DraftAssistThreadMessage[]>>;
  setRunStartedAt: Dispatch<SetStateAction<number | null>>;
};

async function startRunForMessage(
  deps: DraftAssistRunDeps,
  sessionId: string,
  message: string,
  snapshot: DraftAssistSnapshot,
): Promise<void> {
  deps.startAbortRef.current?.abort();
  const controller = new AbortController();
  deps.startAbortRef.current = controller;
  deps.dispatch({ type: "run_requested" });
  deps.setRunStartedAt(Date.now());
  try {
    const result = await apiStartRun(
      sessionId,
      { user_message: message, snapshot },
      { signal: controller.signal },
    );
    deps.runIdRef.current = result.run_id;
    deps.dispatch({ type: "run_accepted" });
  } catch (e: unknown) {
    if (controller.signal.aborted) return;
    const err = e instanceof Error ? e : new Error(String(e));
    deps.dispatch({ type: "transport_error", message: err.message });
    deps.setRunStartedAt(null);
  } finally {
    if (deps.startAbortRef.current === controller) {
      deps.startAbortRef.current = null;
    }
  }
}

function pushUser(deps: DraftAssistRunDeps, text: string): void {
  const at = new Date().toISOString();
  deps.setMessages((prev) => [
    ...prev,
    { id: `user-${prev.length}-${Date.now()}`, kind: "user", text, at },
  ]);
}

/** Open the assist column and optionally start the first run. */
export function draftAssistOpen(
  deps: DraftAssistRunDeps,
  userMessage: string,
): void {
  const trimmed = userMessage.trim();
  deps.setActive(true);
  if (trimmed !== "") pushUser(deps, trimmed);
  deps.assistantTurnIdRef.current = null;
  void (async () => {
    try {
      const snapshot = deps.getSnapshotRef.current();
      const opened = await deps.session.open(snapshot, deps.worktreeId);
      if (trimmed !== "") {
        await startRunForMessage(deps, opened.id, trimmed, snapshot);
      }
    } catch (e: unknown) {
      const err = e instanceof Error ? e : new Error(String(e));
      deps.dispatch({ type: "transport_error", message: err.message });
      deps.setRunStartedAt(null);
    }
  })();
}

/** Send a follow-up on the active session. */
export function draftAssistSend(
  deps: DraftAssistRunDeps,
  userMessage: string,
): void {
  const trimmed = userMessage.trim();
  if (trimmed === "") return;
  deps.setActive(true);
  pushUser(deps, trimmed);
  deps.assistantTurnIdRef.current = null;
  void (async () => {
    try {
      const snapshot = deps.getSnapshotRef.current();
      const opened = await deps.session.open(snapshot, deps.worktreeId);
      await startRunForMessage(deps, opened.id, trimmed, snapshot);
    } catch (e: unknown) {
      const err = e instanceof Error ? e : new Error(String(e));
      deps.dispatch({ type: "transport_error", message: err.message });
      deps.setRunStartedAt(null);
    }
  })();
}

/** Cancel the in-flight run. */
export function draftAssistStop(deps: DraftAssistRunDeps): void {
  const sid = deps.session.session?.id;
  const rid = deps.runIdRef.current;
  deps.dispatch({ type: "cancel_requested" });
  if (!sid || !rid) {
    deps.setRunStartedAt(null);
    return;
  }
  deps.cancelAbortRef.current?.abort();
  const controller = new AbortController();
  deps.cancelAbortRef.current = controller;
  void (async () => {
    try {
      await apiCancelRun(sid, rid, { signal: controller.signal });
    } catch (e: unknown) {
      if (controller.signal.aborted) return;
      const err = e instanceof Error ? e : new Error(String(e));
      deps.dispatch({ type: "transport_error", message: err.message });
    } finally {
      if (deps.cancelAbortRef.current === controller) {
        deps.cancelAbortRef.current = null;
      }
    }
  })();
}

/** Tear down the session and clear local state. */
export function draftAssistReset(deps: DraftAssistRunDeps): void {
  deps.startAbortRef.current?.abort();
  deps.cancelAbortRef.current?.abort();
  deps.session.close();
  deps.setActive(false);
  deps.setMessages([]);
  deps.setRunStartedAt(null);
  deps.runIdRef.current = null;
  deps.assistantTurnIdRef.current = null;
  deps.dispatch({ type: "reset" });
}
