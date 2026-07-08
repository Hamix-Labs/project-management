import { describe, expect, it } from "vitest";
import { isTerminalTaskStatus } from "./terminalTaskStatus";

describe("isTerminalTaskStatus", () => {
  it("returns true for done and failed", () => {
    expect(isTerminalTaskStatus("done")).toBe(true);
    expect(isTerminalTaskStatus("failed")).toBe(true);
  });

  it("returns false for non-terminal statuses", () => {
    for (const status of ["ready", "running", "blocked", "review", "on_hold"] as const) {
      expect(isTerminalTaskStatus(status)).toBe(false);
    }
  });
});
