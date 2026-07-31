import { describe, expect, it } from "vitest";
import type { TaskCyclePhase, TaskCycleStreamEvent } from "@/types";
import { latestAgentReplyByPhase } from "./latestAgentReplyByPhase";

function stream(
  overrides: Partial<TaskCycleStreamEvent> &
    Pick<TaskCycleStreamEvent, "phase_seq" | "stream_seq" | "kind">,
): TaskCycleStreamEvent {
  return {
    id: `ev-${overrides.stream_seq}`,
    task_id: "t1",
    cycle_id: "c1",
    at: "2026-07-30T10:28:00Z",
    source: "cursor",
    payload: {},
    ...overrides,
  };
}

function phase(
  overrides: Partial<TaskCyclePhase> & Pick<TaskCyclePhase, "phase_seq">,
): TaskCyclePhase {
  return {
    id: `p-${overrides.phase_seq}`,
    cycle_id: "c1",
    phase: overrides.phase_seq === 1 ? "execute" : "verify",
    status: "succeeded",
    started_at: "2026-07-30T10:22:00Z",
    details: {},
    ...overrides,
  };
}

describe("latestAgentReplyByPhase", () => {
  it("picks the newest assistant message per phase", () => {
    const events = [
      stream({
        phase_seq: 2,
        stream_seq: 10,
        kind: "assistant",
        message: "older reply",
      }),
      stream({
        phase_seq: 2,
        stream_seq: 20,
        kind: "assistant",
        message: "Verify report submitted",
      }),
      stream({
        phase_seq: 1,
        stream_seq: 5,
        kind: "message",
        message: "Execute done",
      }),
    ];
    const map = latestAgentReplyByPhase(events, [
      phase({ phase_seq: 1 }),
      phase({ phase_seq: 2 }),
    ]);
    expect(map.get(1)).toEqual({
      text: "Execute done",
      at: "2026-07-30T10:28:00Z",
      source: "stream",
    });
    expect(map.get(2)).toEqual({
      text: "Verify report submitted",
      at: "2026-07-30T10:28:00Z",
      source: "stream",
    });
  });

  it("ignores tool_call and tool done events", () => {
    const events = [
      stream({
        phase_seq: 1,
        stream_seq: 2,
        kind: "tool_call",
        message: "Starting hook",
        tool: "hookAdditionalContexts",
      }),
      stream({
        phase_seq: 1,
        stream_seq: 3,
        kind: "tool",
        subtype: "completed",
        message: "Finishing completedAtMs",
        tool: "hookAdditionalContexts",
      }),
      stream({
        phase_seq: 1,
        stream_seq: 1,
        kind: "assistant",
        message: "Only real reply",
      }),
    ];
    const map = latestAgentReplyByPhase(events, [phase({ phase_seq: 1 })]);
    expect(map.get(1)?.text).toBe("Only real reply");
  });

  it("skips empty assistant messages", () => {
    const events = [
      stream({
        phase_seq: 1,
        stream_seq: 2,
        kind: "assistant",
        message: "   ",
      }),
      stream({
        phase_seq: 1,
        stream_seq: 1,
        kind: "assistant",
        message: "kept",
      }),
    ];
    const map = latestAgentReplyByPhase(events, [phase({ phase_seq: 1 })]);
    expect(map.get(1)?.text).toBe("kept");
  });

  it("falls back to phase.summary when stream has no reply", () => {
    const events = [
      stream({
        phase_seq: 1,
        stream_seq: 1,
        kind: "tool_call",
        message: "noise",
      }),
    ];
    const map = latestAgentReplyByPhase(events, [
      phase({ phase_seq: 1, summary: "  All criteria verified  " }),
      phase({ phase_seq: 2 }),
    ]);
    expect(map.get(1)).toEqual({
      text: "All criteria verified",
      source: "summary",
    });
    expect(map.has(2)).toBe(false);
  });

  it("prefers last stream agent reply over phase.summary", () => {
    const events = [
      stream({
        phase_seq: 1,
        stream_seq: 10,
        kind: "assistant",
        message: "I've committed the refactor and verified the commands.",
      }),
      stream({
        phase_seq: 1,
        stream_seq: 40,
        kind: "assistant",
        message: "Refactor is complete and committed.",
      }),
    ];
    const map = latestAgentReplyByPhase(events, [
      phase({
        phase_seq: 1,
        summary: "A longer terminal result.result that is not the last reply.",
      }),
    ]);
    expect(map.get(1)?.source).toBe("stream");
    expect(map.get(1)?.text).toBe("Refactor is complete and committed.");
  });

  it("recovers full assistant text from payload when message was clipped", () => {
    const full =
      "Refactor is complete and committed.\n\n" +
      "- Longest eligible function identified and refactored.\n" +
      "- Extracted descriptive helpers for the remaining workflow steps.";
    const clipped = full.slice(0, 80) + "…";
    const events = [
      stream({
        phase_seq: 1,
        stream_seq: 1,
        kind: "assistant",
        message: clipped,
        payload: {
          type: "assistant",
          message: {
            role: "assistant",
            content: [{ type: "text", text: full }],
          },
        },
      }),
    ];
    const map = latestAgentReplyByPhase(events, [phase({ phase_seq: 1 })]);
    expect(map.get(1)?.text).toBe(full);
  });

  it("keeps message when payload text is not longer", () => {
    const events = [
      stream({
        phase_seq: 1,
        stream_seq: 1,
        kind: "assistant",
        message: "short reply",
        payload: {
          type: "assistant",
          message: {
            role: "assistant",
            content: [{ type: "text", text: "short" }],
          },
        },
      }),
    ];
    const map = latestAgentReplyByPhase(events, [phase({ phase_seq: 1 })]);
    expect(map.get(1)?.text).toBe("short reply");
  });

  it("keeps last stream reply even when clipped and summary is longer", () => {
    const events = [
      stream({
        phase_seq: 1,
        stream_seq: 99,
        kind: "assistant",
        message: "Refactor is complete and committed.\n\n- Extracted descriptive …",
        payload: {},
      }),
    ];
    const map = latestAgentReplyByPhase(events, [
      phase({
        phase_seq: 1,
        summary:
          "A much longer phase.summary from result.result that operators should not see when a stream reply exists.",
      }),
    ]);
    expect(map.get(1)?.source).toBe("stream");
    expect(map.get(1)?.text).toBe(
      "Refactor is complete and committed.\n\n- Extracted descriptive …",
    );
  });

  it("reads harness ProgressEvent fallback payload.message string", () => {
    const events = [
      stream({
        phase_seq: 1,
        stream_seq: 1,
        kind: "assistant",
        message: "clipped…",
        payload: {
          Kind: "assistant",
          Message: "should ignore Go casing",
          message: "full recovered text from harness fallback",
        },
      }),
    ];
    const map = latestAgentReplyByPhase(events, [phase({ phase_seq: 1 })]);
    expect(map.get(1)?.text).toBe("full recovered text from harness fallback");
  });
});
