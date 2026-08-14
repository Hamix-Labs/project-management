import type { MutableRefObject } from "react";
import type { DraftAssistThreadMessage } from "./draftAssistTypes";

/** Append a streaming token onto the current assistant turn (or start one). */
export function appendAssistantToken(
  prev: DraftAssistThreadMessage[],
  at: string,
  delta: string,
  turnRef: MutableRefObject<string | null>,
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
  return [...prev, { id, kind: "assistant", text: delta, at, done: false }];
}

/** Mark the current assistant turn done so the next Send starts a new bubble. */
export function closeAssistantTurn(
  prev: DraftAssistThreadMessage[],
  turnRef: MutableRefObject<string | null>,
): DraftAssistThreadMessage[] {
  const id = turnRef.current;
  if (!id) return prev;
  return prev.map((m) =>
    m.id === id && m.kind === "assistant" ? { ...m, done: true } : m,
  );
}
