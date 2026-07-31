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

function looksClipped(text: string): boolean {
  return text.endsWith("…") || text.endsWith("...");
}

/**
 * Extracts assistant text from a Cursor-shaped stream payload
 * (`message.content[].text`) or a harness ProgressEvent fallback
 * (`message` string). Returns undefined when neither shape has text.
 */
export function assistantTextFromPayload(
  payload: Record<string, unknown> | undefined,
): string | undefined {
  if (!payload || typeof payload !== "object") return undefined;

  const message = payload.message;
  if (typeof message === "string") {
    const trimmed = message.trim();
    return trimmed || undefined;
  }
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
      parts.push(rec.text.trim());
    }
  }
  // Match Go textContent: join text parts with newlines.
  const joined = parts.join("\n").trim();
  return joined || undefined;
}

/**
 * Prefers payload text when it is longer than the stored message, or when the
 * message looks clipped (legacy 240-rune + ellipsis rows).
 */
export function resolveAgentReplyText(ev: TaskCycleStreamEvent): string | undefined {
  const message = ev.message?.trim();
  const fromPayload = assistantTextFromPayload(ev.payload);
  if (!fromPayload) return message || undefined;
  if (!message) return fromPayload;
  if (fromPayload.length > message.length) return fromPayload;
  if (looksClipped(message) && fromPayload.length >= message.length - 1) {
    return fromPayload;
  }
  return message;
}

/**
 * Picks the best Agent reply per phase: newest stream assistant/message text
 * (with payload recovery), then phase.summary when it is longer or the stream
 * text looks clipped. Empty messages are ignored.
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
    const summary = phase.summary?.trim();
    if (fromStream) {
      const preferSummary =
        !!summary &&
        (summary.length > fromStream.text.length ||
          (looksClipped(fromStream.text) &&
            summary.length >= fromStream.text.length - 1));
      if (preferSummary && summary) {
        out.set(phase.phase_seq, { text: summary, source: "summary" });
        continue;
      }
      out.set(phase.phase_seq, {
        text: fromStream.text,
        at: fromStream.at,
        source: "stream",
      });
      continue;
    }
    if (summary) {
      out.set(phase.phase_seq, { text: summary, source: "summary" });
    }
  }
  return out;
}
