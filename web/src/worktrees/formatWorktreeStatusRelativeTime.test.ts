import { describe, expect, it } from "vitest";
import { formatWorktreeStatusRelativeTime } from "./formatWorktreeStatusRelativeTime";

const NOW = new Date("2026-07-04T18:00:00.000Z");

function isoMinus(ms: number): string {
  return new Date(NOW.getTime() - ms).toISOString();
}

describe("formatWorktreeStatusRelativeTime", () => {
  it("uses prose minutes and hours", () => {
    expect(formatWorktreeStatusRelativeTime(isoMinus(12 * 60_000), NOW)).toBe(
      "12 minutes ago",
    );
    expect(formatWorktreeStatusRelativeTime(isoMinus(2 * 3_600_000), NOW)).toBe(
      "2 hours ago",
    );
  });

  it("uses yesterday for one day", () => {
    expect(formatWorktreeStatusRelativeTime(isoMinus(86_400_000), NOW)).toBe("yesterday");
  });
});
