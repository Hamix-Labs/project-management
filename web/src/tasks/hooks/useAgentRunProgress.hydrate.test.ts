import { describe, expect, it } from "vitest";
import {
  hydrateAgentRunProgress,
  pushAgentRunProgress,
  resetAgentRunProgressForTests,
  useAgentRunProgress,
} from "./useAgentRunProgress";
import { renderHook } from "@testing-library/react";

describe("hydrateAgentRunProgress", () => {
  it("seeds empty phase from durable stream events", () => {
    resetAgentRunProgressForTests();
    hydrateAgentRunProgress("t1", "c1", 1, [
      {
        kind: "tool_call",
        subtype: "started",
        tool: "Shell",
        message: "Run tests",
        at: "2026-07-30T17:24:00Z",
        phase_seq: 1,
      },
    ]);
    const { result } = renderHook(() => useAgentRunProgress("t1", "c1", 1));
    expect(result.current).toHaveLength(1);
    expect(result.current[0]?.progress.tool).toBe("Shell");
  });

  it("does not overwrite existing live SSE items", () => {
    resetAgentRunProgressForTests();
    pushAgentRunProgress({
      taskId: "t1",
      cycleId: "c1",
      phaseSeq: 1,
      progress: { kind: "assistant", message: "live" },
    });
    hydrateAgentRunProgress("t1", "c1", 1, [
      { kind: "tool_call", tool: "Shell", phase_seq: 1 },
    ]);
    const { result } = renderHook(() => useAgentRunProgress("t1", "c1", 1));
    expect(result.current).toHaveLength(1);
    expect(result.current[0]?.progress.message).toBe("live");
  });
});
