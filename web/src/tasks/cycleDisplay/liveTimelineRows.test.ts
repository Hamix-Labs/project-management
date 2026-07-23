import { describe, expect, it } from "vitest";
import type { AgentRunProgressItem } from "@/tasks/hooks/useAgentRunProgress";
import {
  liveTimelineIconForTone,
  toLiveTimelineRows,
} from "./liveTimelineRows";

const NOW = 1_000_000;

function frame(
  overrides: Partial<AgentRunProgressItem> &
    Pick<AgentRunProgressItem, "progress">,
  receivedAt: number,
): AgentRunProgressItem {
  return {
    taskId: "task-1",
    cycleId: "cycle-1",
    phaseSeq: 1,
    receivedAt,
    ...overrides,
  };
}

describe("liveTimelineIconForTone", () => {
  it("maps tones to icon roles", () => {
    expect(liveTimelineIconForTone("done")).toBe("done");
    expect(liveTimelineIconForTone("tool")).toBe("call");
    expect(liveTimelineIconForTone("failed")).toBe("failed");
    expect(liveTimelineIconForTone("error")).toBe("failed");
    expect(liveTimelineIconForTone("reply")).toBe("neutral");
    expect(liveTimelineIconForTone("session")).toBe("neutral");
    expect(liveTimelineIconForTone("neutral")).toBe("neutral");
  });
});

describe("toLiveTimelineRows", () => {
  it("returns empty when there are no items", () => {
    expect(toLiveTimelineRows([], NOW)).toEqual([]);
  });

  it("prepends a Working row with relative time (no Last prefix)", () => {
    const items = [
      frame(
        {
          progress: {
            kind: "tool_call",
            subtype: "completed",
            message: "Done reading",
          },
        },
        NOW - 2_000,
      ),
    ];
    const rows = toLiveTimelineRows(items, NOW, {
      showPendingRow: true,
      pendingMessage: "Waiting for the next agent update…",
    });
    expect(rows[0]).toMatchObject({
      isPending: true,
      icon: "working",
      kindLabel: "Working",
      message: "Waiting for the next agent update…",
      messageEmphasis: "secondary",
      timeLabel: "2s ago",
    });
    expect(rows[0].timeLabel).not.toMatch(/^Last /);
    expect(rows[1]).toMatchObject({
      icon: "done",
      kindLabel: "Tool done",
      message: "Done reading",
      messageEmphasis: "primary",
    });
  });

  it("caps non-pending rows at maxItems and sorts newest first", () => {
    const items = [
      frame({ progress: { kind: "system", message: "older" } }, NOW - 20_000),
      frame({ progress: { kind: "assistant", message: "newest" } }, NOW - 1_000),
      frame(
        { progress: { kind: "tool_call", message: "mid" } },
        NOW - 10_000,
      ),
      frame({ progress: { kind: "system", message: "oldest" } }, NOW - 30_000),
    ];
    const rows = toLiveTimelineRows(items, NOW, {
      maxItems: 2,
      showPendingRow: false,
    });
    expect(rows).toHaveLength(2);
    expect(rows[0].message).toBe("newest");
    expect(rows[1].message).toBe("mid");
    expect(rows[1].icon).toBe("call");
  });

  it("maps tool call / failed tones onto call and failed icons", () => {
    const items = [
      frame(
        {
          progress: {
            kind: "tool_call",
            subtype: "failed",
            message: "boom",
          },
        },
        NOW - 1_000,
      ),
      frame(
        {
          progress: {
            kind: "tool_call",
            message: "starting",
          },
        },
        NOW - 5_000,
      ),
    ];
    const rows = toLiveTimelineRows(items, NOW, { showPendingRow: false });
    expect(rows[0].icon).toBe("failed");
    expect(rows[1].icon).toBe("call");
  });
});
