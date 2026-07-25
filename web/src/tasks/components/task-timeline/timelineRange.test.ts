import { describe, expect, it } from "vitest";
import {
  eventInTimelineRange,
  timelineRangeCutoff,
  timelineRangeLabel,
} from "./timelineRange";

describe("timelineRange", () => {
  const now = new Date("2026-07-25T18:00:00.000Z");

  it("labels known range ids", () => {
    expect(timelineRangeLabel("7d")).toBe("Last 7 days");
    expect(timelineRangeLabel("all")).toBe("All time");
  });

  it("returns null cutoff for all time", () => {
    expect(timelineRangeCutoff("all", now)).toBeNull();
  });

  it("filters by lookback window", () => {
    expect(
      eventInTimelineRange("2026-07-25T12:00:00.000Z", "24h", now),
    ).toBe(true);
    expect(
      eventInTimelineRange("2026-07-23T12:00:00.000Z", "24h", now),
    ).toBe(false);
    expect(
      eventInTimelineRange("2026-07-20T12:00:00.000Z", "7d", now),
    ).toBe(true);
    expect(
      eventInTimelineRange("2026-06-01T12:00:00.000Z", "all", now),
    ).toBe(true);
  });
});
