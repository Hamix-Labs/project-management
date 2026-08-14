import { useCallback, useEffect, useMemo, useReducer, useRef, useState } from "react";
import type { DraftAssistSnapshot } from "@/types/draftAssist";
import { useDraftAssistSession } from "@/tasks/hooks/useDraftAssistSession";
import { useDraftAssistStream } from "@/tasks/hooks/useDraftAssistStream";
import {
  draftAssistStatusReducer,
  INITIAL_DRAFT_ASSIST_STATUS,
  isDraftAssistRunActive,
} from "./draftAssistStatus";
import { handleDraftAssistEvent } from "./draftAssistEventHandler";
import {
  draftAssistOpen,
  draftAssistReset,
  draftAssistSend,
  draftAssistStop,
  type DraftAssistRunDeps,
} from "./draftAssistRunActions";
import type {
  DraftAssistContextValue,
  DraftAssistThreadMessage,
} from "./draftAssistTypes";

type Args = {
  getSnapshot: () => DraftAssistSnapshot;
  worktreeId?: string;
  onApplyPromptPatch?: (nextPrompt: string) => void;
  getPromptSnapshot?: () => string;
};

/** Session, SSE, status machine, and message-list controller for draft assist. */
export function useDraftAssistController({
  getSnapshot,
  worktreeId,
  onApplyPromptPatch,
  getPromptSnapshot,
}: Args): DraftAssistContextValue {
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

  const handleEvent = useCallback(
    (event: Parameters<typeof handleDraftAssistEvent>[0]) => {
      handleDraftAssistEvent(event, dispatch, setMessages, setRunStartedAt, {
        assistantTurnIdRef,
        runIdRef,
        getPromptRef,
        applyPromptRef,
      });
    },
    [],
  );

  const stream = useDraftAssistStream(session.session?.id ?? null, {
    enabled: active && session.session != null,
    onEvent: handleEvent,
  });

  useEffect(() => {
    if (!active) return;
    if (stream.status === "disconnected") {
      dispatch({ type: "connection", connected: false });
    } else if (stream.status === "live") {
      dispatch({ type: "connection", connected: true });
    }
  }, [stream.status, active]);

  const deps = useMemo<DraftAssistRunDeps>(
    () => ({
      session,
      worktreeId,
      getSnapshotRef,
      runIdRef,
      assistantTurnIdRef,
      startAbortRef,
      cancelAbortRef,
      dispatch,
      setActive,
      setMessages,
      setRunStartedAt,
    }),
    [session, worktreeId],
  );

  const open = useCallback(
    (msg: string) => draftAssistOpen(deps, msg),
    [deps],
  );
  const send = useCallback(
    (msg: string) => draftAssistSend(deps, msg),
    [deps],
  );
  const stop = useCallback(() => draftAssistStop(deps), [deps]);
  const reset = useCallback(() => draftAssistReset(deps), [deps]);

  useEffect(
    () => () => {
      startAbortRef.current?.abort();
      cancelAbortRef.current?.abort();
    },
    [],
  );

  return {
    active,
    status,
    messages,
    sessionId: session.session?.id ?? null,
    runStartedAt,
    runActive: isDraftAssistRunActive(status),
    open,
    send,
    stop,
    reset,
  };
}
