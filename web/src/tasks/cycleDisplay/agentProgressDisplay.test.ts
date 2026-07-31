import { describe, expect, it } from "vitest";
import {
  agentProgressKindDescriptor,
  agentProgressKindLabel,
  agentProgressMessage,
  formatAgentProgressElapsed,
} from "./agentProgressDisplay";
import type { AgentRunProgressItem } from "@/tasks/hooks/useAgentRunProgress";

function item(
  overrides: Partial<AgentRunProgressItem> & Pick<AgentRunProgressItem, "progress">,
): AgentRunProgressItem {
  return {
    taskId: "task-1",
    cycleId: "cycle-1",
    phaseSeq: 1,
    receivedAt: 1_000_000,
    ...overrides,
  };
}

describe("agentProgressDisplay", () => {
  it.each([
    ["tool_call", "done", undefined, "Tool done"],
    ["tool_call", "error", "Read", "Tool failed"],
    ["tool_call", "started", "Read", "Tool call"],
    ["tool_call", "started", "verify_command", "Verify command"],
    ["tool_call", "running", "verify_command", "Verify command"],
    ["tool_call", "completed", "verify_command", "Command done"],
    ["tool_call", "failed", "verify_command", "Command failed"],
    ["assistant", undefined, undefined, "Agent reply"],
    ["run_state", "setup_started", "harness_setup", "Setup"],
    ["run_state", "setup_spawn", "harness_setup", "Setup"],
    ["run_state", "handoff_claims", "harness_setup", "Setup"],
    ["run_state", "handoff_verify", "harness_setup", "Setup"],
    ["run_state", "restart_resume", "harness_setup", "Setup"],
    ["custom_event", undefined, undefined, "custom event"],
  ] as const)(
    "agentProgressKindLabel(%s, %s, %s) → %s",
    (kind, subtype, tool, expected) => {
      expect(agentProgressKindLabel(kind, subtype, tool)).toBe(expected);
    },
  );

  it.each([
    [
      "started",
      "Execute agent is running a checklist verify command",
    ],
    ["completed", "Execute agent verify command finished"],
    ["failed", "Execute agent verify command failed"],
  ] as const)(
    "verify_command title for subtype %s names the execute agent",
    (subtype, title) => {
      expect(
        agentProgressKindDescriptor("tool_call", subtype, "verify_command").title,
      ).toBe(title);
    },
  );

  it("agentProgressMessage prefers message then tool then fallback", () => {
    expect(
      agentProgressMessage(
        item({ progress: { kind: "assistant", message: "Hello" } }),
      ),
    ).toBe("Hello");
    expect(
      agentProgressMessage(item({ progress: { kind: "tool_call", tool: "Grep" } })),
    ).toBe("Grep");
    expect(
      agentProgressMessage(item({ progress: { kind: "system" } })),
    ).toBe("Working…");
  });

  it.each(["handoff_claims", "handoff_verify"] as const)(
    "agentProgressMessage remaps %s to claim acceptance copy",
    (subtype) => {
      expect(
        agentProgressMessage(
          item({
            progress: {
              kind: "run_state",
              subtype,
              message: "Handing off to verify…",
              tool: "harness_setup",
            },
          }),
        ),
      ).toBe("Accepting criteria claims…");
    },
  );
  it.each([
    [1_000_000, 1_000_000, "just now"],
    [1_000_000, 1_000_500, "just now"],
    [1_000_000, 1_012_000, "12s ago"],
    [1_000_000, 1_130_000, "2m ago"],
    [1_000_000, 999_000, "just now"],
  ] as const)(
    "formatAgentProgressElapsed handles elapsed edge (%i, %i)",
    (receivedAt, now, expected) => {
      expect(formatAgentProgressElapsed(receivedAt, now)).toBe(expected);
    },
  );
});
