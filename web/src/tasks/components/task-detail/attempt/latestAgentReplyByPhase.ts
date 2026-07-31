import type { TaskCyclePhase, TaskCycleStreamEvent } from "@/types";

export type PhaseAgentReplySource = "stream" | "summary";

export type PhaseAgentReply = {
  text: string;
  at?: string;
  source: PhaseAgentReplySource;
};

function isAgentReplyKind(kind: string): boolean {
  return kind === "assistant" || kind === "message";
}

/**
 * Extracts assistant text from a Cursor-shaped stream payload
 * (`message.content[].text`). Returns undefined when the payload is not shaped
 * that way or has no text parts.
 */
export function assistantTextFromPayload(
  payload: Record<string, unknown> | undefined,
): string | undefined {
  if (!payload || typeof payload !== "object") return undefined;
  const message = payload.message;
  if (!message || typeof message !== "object" || Array.isArray(message)) {
    return undefined;
  }
  const content = (message as { content?: unknown }).content;
  if (!Array.isArray(content)) return undefined;
  const parts: string[] = [];
  for (const item of content) {
    if (!item || typeof item !== "object" || Array.isArray(item)) continue;
    const rec = item as { type?: unknown; text?: unknown };
    if (rec.type === "text" && typeof rec.text === "string" && rec.text.trim()) {
      parts.push(rec.text);
    }
  }
  const joined = parts.join("").trim();
  return joined || undefined;
}

/**
 * Prefers payload text when it is longer than the stored message so View reply
 * recovers full content from rows clipped at persist time (240-rune cap).
 */
export function resolveAgentReplyText(ev: TaskCycleStreamEvent): string | undefined {
  const message = ev.message?.trim();
  const fromPayload = assistantTextFromPayload(ev.payload);
  if (fromPayload && (!message || fromPayload.length > message.length)) {
    return fromPayload;
  }
  return message || undefined;
}

/**
 * Picks the newest Agent reply (assistant/message stream event) per phase.
 * Falls back to `phase.summary` when the loaded stream has no reply for that
 * phase. Empty messages are ignored.
 */
export function latestAgentReplyByPhase(
  events: readonly TaskCycleStreamEvent[],
  phases: readonly TaskCyclePhase[],
): Map<number, PhaseAgentReply> {
  const bestByPhase = new Map<
    number,
    { text: string; at: string; streamSeq: number }
  >();

  for (const ev of events) {
    if (!isAgentReplyKind(ev.kind)) continue;
    const text = resolveAgentReplyText(ev);
    if (!text) continue;
    const prev = bestByPhase.get(ev.phase_seq);
    if (prev && prev.streamSeq >= ev.stream_seq) continue;
    bestByPhase.set(ev.phase_seq, {
      text,
      at: ev.at,
      streamSeq: ev.stream_seq,
    });
  }

  const out = new Map<number, PhaseAgentReply>();
  for (const phase of phases) {
    const fromStream = bestByPhase.get(phase.phase_seq);
    if (fromStream) {
      out.set(phase.phase_seq, {
        text: fromStream.text,
        at: fromStream.at,
        source: "stream",
      });
      continue;
    }
    const summary = phase.summary?.trim();
    if (summary) {
      out.set(phase.phase_seq, { text: summary, source: "summary" });
    }
  }
  return out;
}
