import { describe, expect, it } from "vitest";
import {
  earliestCycleStartedAt,
  formatTaskCompletionDuration,
} from "./taskCompletionDuration";

describe("earliestCycleStartedAt", () => {
  it("returns the earliest started_at among cycles", () => {
    expect(
      earliestCycleStartedAt([
        { started_at: "2026-04-22T14:00:00Z" },
        { started_at: "2026-04-22T13:00:00Z" },
        { started_at: "2026-04-22T13:30:00Z" },
      ]),
    ).toBe("2026-04-22T13:00:00Z");
  });

  it("returns null when there are no usable starts", () => {
    expect(earliestCycleStartedAt([])).toBeNull();
    expect(earliestCycleStartedAt([{ started_at: "" }])).toBeNull();
  });
});

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
