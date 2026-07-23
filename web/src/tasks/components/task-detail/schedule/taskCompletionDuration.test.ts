import { describe, expect, it } from "vitest";
import { formatTaskCompletionDuration } from "./taskCompletionDuration";

describe("formatTaskCompletionDuration", () => {
  it("formats a multi-minute span", () => {
    expect(
      formatTaskCompletionDuration(
        "2026-04-22T13:00:00Z",
        "2026-04-22T13:12:00Z",
      ),
    ).toBe("12 min");
  });

  it("returns null for missing or inverted times", () => {
    expect(formatTaskCompletionDuration(null, "2026-04-22T13:00:00Z")).toBeNull();
    expect(formatTaskCompletionDuration("2026-04-22T14:00:00Z", "2026-04-22T13:00:00Z")).toBeNull();
  });
});
