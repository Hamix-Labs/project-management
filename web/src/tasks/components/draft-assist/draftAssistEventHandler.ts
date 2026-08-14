import type { Dispatch, MutableRefObject, SetStateAction } from "react";
import type { DraftAssistEvent } from "@/types/draftAssist";
import { applyDraftAssistPatch } from "./draftAssistPatch";
import {
  appendAssistantToken,
  closeAssistantTurn,
} from "./draftAssistMessages";
import type { DraftAssistStatusAction } from "./draftAssistStatus";
import type { DraftAssistThreadMessage } from "./draftAssistTypes";

type Refs = {
  assistantTurnIdRef: MutableRefObject<string | null>;
  runIdRef: MutableRefObject<string | null>;
  getPromptRef: MutableRefObject<(() => string) | undefined>;
  applyPromptRef: MutableRefObject<((next: string) => void) | undefined>;
};

/** Apply one SSE frame to status + message list. */
export function handleDraftAssistEvent(
  event: DraftAssistEvent,
  dispatch: Dispatch<DraftAssistStatusAction>,
  setMessages: Dispatch<SetStateAction<DraftAssistThreadMessage[]>>,
  setRunStartedAt: Dispatch<SetStateAction<number | null>>,
  refs: Refs,
): void {
  dispatch({ type: "event", event });
  switch (event.kind) {
    case "session":
    case "status":
      return;
    case "token":
      setMessages((prev) =>
        appendAssistantToken(
          prev,
          event.at,
          event.data.delta,
          refs.assistantTurnIdRef,
        ),
      );
      return;
    case "tool":
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
    case "patch": {
      let applied = false;
      try {
        const current = refs.getPromptRef.current?.() ?? "";
        const next = applyDraftAssistPatch(current, event.data);
        if (next !== null && refs.applyPromptRef.current) {
          refs.applyPromptRef.current(next);
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
    case "error":
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
    case "done":
      setMessages((prev) => closeAssistantTurn(prev, refs.assistantTurnIdRef));
      refs.assistantTurnIdRef.current = null;
      refs.runIdRef.current = null;
      setRunStartedAt(null);
      return;
  }
}
