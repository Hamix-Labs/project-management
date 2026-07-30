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
    const text = ev.message?.trim();
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
