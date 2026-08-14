import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
  useRef,
  useState,
  type ReactNode,
} from "react";
import {
  cancelRun as apiCancelRun,
  startRun as apiStartRun,
} from "@/api/draftAssist";
import type {
  DraftAssistEvent,
  DraftAssistPatchOp,
  DraftAssistSnapshot,
} from "@/types/draftAssist";
import { useDraftAssistSession } from "@/tasks/hooks/useDraftAssistSession";
import { useDraftAssistStream } from "@/tasks/hooks/useDraftAssistStream";
import {
  draftAssistStatusReducer,
  INITIAL_DRAFT_ASSIST_STATUS,
  isDraftAssistRunActive,
  type DraftAssistStatusState,
} from "./draftAssistStatus";
import { applyDraftAssistPatch } from "./draftAssistPatch";

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

const DraftAssistCtx = createContext<DraftAssistContextValue | null>(null);

/** Consumer hook — throws when called outside a {@link DraftAssistProvider}. */
export function useDraftAssistContext(): DraftAssistContextValue {
  const ctx = useContext(DraftAssistCtx);
  if (!ctx) {
    throw new Error(
      "useDraftAssistContext must be rendered inside DraftAssistProvider",
    );
  }
  return ctx;
}

/** Optional consumer hook that returns null when no provider is above. */
export function useOptionalDraftAssistContext(): DraftAssistContextValue | null {
  return useContext(DraftAssistCtx);
}

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

/**
 * Provides the assist session, SSE stream, status machine, and message
 * list. Session creation is lazy: the first call to {@link
 * DraftAssistContextValue.open} POSTs `/sessions` and starts a run;
 * subsequent {@link DraftAssistContextValue.send} calls reuse the
 * cached session.
 *
 * Patches are applied through `onApplyPromptPatch` so the TipTap
 * editor updates in place via the compose form's controlled prompt
 * state — no editor ref plumbing required.
 */
export function DraftAssistProvider({
  getSnapshot,
  worktreeId,
  onApplyPromptPatch,
  getPromptSnapshot,
  children,
}: DraftAssistProviderProps) {
  const [active, setActive] = useState(false);
  const [messages, setMessages] = useState<DraftAssistThreadMessage[]>([]);
  const [status, dispatch] = useReducer(
    draftAssistStatusReducer,
    INITIAL_DRAFT_ASSIST_STATUS,
  );
  const [runStartedAt, setRunStartedAt] = useState<number | null>(null);
  const runIdRef = useRef<string | null>(null);
  const assistantTurnIdRef = useRef<string | null>(null);
  const cancelAbortRef = useRef<AbortController | null>(null);
  const startAbortRef = useRef<AbortController | null>(null);

  const session = useDraftAssistSession();

  // Keep the latest snapshot/prompt readers in refs so callback identity
  // (`open`/`send`) stays stable across re-renders. Callers may re-render
  // on every keystroke; we do not want to churn subscriptions.
  const getSnapshotRef = useRef(getSnapshot);
  useEffect(() => {
    getSnapshotRef.current = getSnapshot;
  }, [getSnapshot]);
  const getPromptRef = useRef(getPromptSnapshot);
  useEffect(() => {
    getPromptRef.current = getPromptSnapshot;
  }, [getPromptSnapshot]);
  const applyPromptRef = useRef(onApplyPromptPatch);
  useEffect(() => {
    applyPromptRef.current = onApplyPromptPatch;
  }, [onApplyPromptPatch]);

  const handleEvent = useCallback((event: DraftAssistEvent) => {
    dispatch({ type: "event", event });

    switch (event.kind) {
      case "session":
        // Handshake only — nothing to append.
        return;

      case "status":
        // Reducer already reflected it; no thread row.
        return;

      case "token": {
        const delta = event.data.delta;
        setMessages((prev) => appendAssistantToken(prev, event.at, delta, assistantTurnIdRef));
        return;
      }

      case "tool": {
        setMessages((prev) => [
          ...prev,
          {
            id: `tool-${event.id}`,
            kind: "tool",
            name: event.data.name,
            phase: event.data.phase,
            ok: event.data.ok,
            error: event.data.error,
            at: event.at,
          },
        ]);
        return;
      }

      case "patch": {
        let applied = false;
        try {
          const current = getPromptRef.current?.() ?? "";
          const next = applyDraftAssistPatch(current, event.data);
          if (next !== null && applyPromptRef.current) {
            applyPromptRef.current(next);
            applied = true;
          }
        } catch {
          applied = false;
        }
        setMessages((prev) => [
          ...prev,
          {
            id: `patch-${event.id}`,
            kind: "patch",
            op: event.data.op,
            summary: event.data.summary,
            applied,
            at: event.at,
          },
        ]);
        return;
      }

      case "error": {
        setMessages((prev) => [
          ...prev,
          {
            id: `err-${event.id}`,
            kind: "error",
            message: event.data.message,
            at: event.at,
          },
        ]);
        return;
      }

      case "done": {
        // Close the current assistant turn so the next Send starts a new bubble.
        setMessages((prev) => closeAssistantTurn(prev, assistantTurnIdRef));
        assistantTurnIdRef.current = null;
        runIdRef.current = null;
        setRunStartedAt(null);
        return;
      }
    }
  }, []);

  const stream = useDraftAssistStream(session.session?.id ?? null, {
    enabled: active && session.session != null,
    onEvent: handleEvent,
  });

  // Reflect transport status → reducer so the disconnected/reconnected copy
  // is centralized in one place.
  useEffect(() => {
    if (!active) return;
    if (stream.status === "disconnected") {
      dispatch({ type: "connection", connected: false });
    } else if (stream.status === "live") {
      dispatch({ type: "connection", connected: true });
    }
  }, [stream.status, active]);

  const startRunForMessage = useCallback(
    async (
      sessionId: string,
      message: string,
      snapshot: DraftAssistSnapshot,
    ) => {
      startAbortRef.current?.abort();
      const controller = new AbortController();
      startAbortRef.current = controller;
      dispatch({ type: "run_requested" });
      setRunStartedAt(Date.now());
      try {
        const result = await apiStartRun(
          sessionId,
          { user_message: message, snapshot },
          { signal: controller.signal },
        );
        runIdRef.current = result.run_id;
        dispatch({ type: "run_accepted" });
      } catch (e: unknown) {
        if (controller.signal.aborted) return;
        const err = e instanceof Error ? e : new Error(String(e));
        dispatch({ type: "transport_error", message: err.message });
        setRunStartedAt(null);
      } finally {
        if (startAbortRef.current === controller) {
          startAbortRef.current = null;
        }
      }
    },
    [],
  );

  const open = useCallback(
    (userMessage: string) => {
      const trimmed = userMessage.trim();
      const at = new Date().toISOString();
      setActive(true);
      // Optimistic user bubble — synchronous per design (<100ms Send→Thinking).
      if (trimmed !== "") {
        setMessages((prev) => [
          ...prev,
          {
            id: `user-${prev.length}-${Date.now()}`,
            kind: "user",
            text: trimmed,
            at,
          },
        ]);
      }
      assistantTurnIdRef.current = null;

      void (async () => {
        try {
          const snapshot = getSnapshotRef.current();
          const opened = await session.open(snapshot, worktreeId);
          if (trimmed !== "") {
            await startRunForMessage(opened.id, trimmed, snapshot);
          }
        } catch (e: unknown) {
          const err = e instanceof Error ? e : new Error(String(e));
          dispatch({ type: "transport_error", message: err.message });
          setRunStartedAt(null);
        }
      })();
    },
    [session, startRunForMessage, worktreeId],
  );

  const send = useCallback(
    (userMessage: string) => {
      const trimmed = userMessage.trim();
      if (trimmed === "") return;
      const at = new Date().toISOString();
      setActive(true);
      setMessages((prev) => [
        ...prev,
        {
          id: `user-${prev.length}-${Date.now()}`,
          kind: "user",
          text: trimmed,
          at,
        },
      ]);
      assistantTurnIdRef.current = null;

      void (async () => {
        try {
          const snapshot = getSnapshotRef.current();
          const opened = await session.open(snapshot, worktreeId);
          await startRunForMessage(opened.id, trimmed, snapshot);
        } catch (e: unknown) {
          const err = e instanceof Error ? e : new Error(String(e));
          dispatch({ type: "transport_error", message: err.message });
          setRunStartedAt(null);
        }
      })();
    },
    [session, startRunForMessage, worktreeId],
  );

  const stop = useCallback(() => {
    const sid = session.session?.id;
    const rid = runIdRef.current;
    if (!sid || !rid) {
      // Nothing to cancel server-side; still surface intent locally so the
      // watchdog stops nagging.
      dispatch({ type: "cancel_requested" });
      setRunStartedAt(null);
      return;
    }
    dispatch({ type: "cancel_requested" });
    cancelAbortRef.current?.abort();
    const controller = new AbortController();
    cancelAbortRef.current = controller;
    void (async () => {
      try {
        await apiCancelRun(sid, rid, { signal: controller.signal });
      } catch (e: unknown) {
        if (controller.signal.aborted) return;
        const err = e instanceof Error ? e : new Error(String(e));
        dispatch({ type: "transport_error", message: err.message });
      } finally {
        if (cancelAbortRef.current === controller) {
          cancelAbortRef.current = null;
        }
      }
    })();
  }, [session.session?.id]);

  const reset = useCallback(() => {
    startAbortRef.current?.abort();
    cancelAbortRef.current?.abort();
    session.close();
    setActive(false);
    setMessages([]);
    setRunStartedAt(null);
    runIdRef.current = null;
    assistantTurnIdRef.current = null;
    dispatch({ type: "reset" });
  }, [session]);

  useEffect(() => {
    return () => {
      startAbortRef.current?.abort();
      cancelAbortRef.current?.abort();
    };
  }, []);

  const runActive = isDraftAssistRunActive(status);

  const value = useMemo<DraftAssistContextValue>(
    () => ({
      active,
      status,
      messages,
      sessionId: session.session?.id ?? null,
      runStartedAt,
      runActive,
      open,
      send,
      stop,
      reset,
    }),
    [
      active,
      status,
      messages,
      session.session?.id,
      runStartedAt,
      runActive,
      open,
      send,
      stop,
      reset,
    ],
  );

  return (
    <DraftAssistCtx.Provider value={value}>{children}</DraftAssistCtx.Provider>
  );
}

function appendAssistantToken(
  prev: DraftAssistThreadMessage[],
  at: string,
  delta: string,
  turnRef: React.MutableRefObject<string | null>,
): DraftAssistThreadMessage[] {
  if (turnRef.current) {
    return prev.map((m) =>
      m.id === turnRef.current && m.kind === "assistant"
        ? { ...m, text: m.text + delta, at }
        : m,
    );
  }
  const id = `assistant-${prev.length}-${Date.now()}`;
  turnRef.current = id;
  return [
    ...prev,
    { id, kind: "assistant", text: delta, at, done: false },
  ];
}

function closeAssistantTurn(
  prev: DraftAssistThreadMessage[],
  turnRef: React.MutableRefObject<string | null>,
): DraftAssistThreadMessage[] {
  const id = turnRef.current;
  if (!id) return prev;
  return prev.map((m) =>
    m.id === id && m.kind === "assistant" ? { ...m, done: true } : m,
  );
}
