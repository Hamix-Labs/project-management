import { describe, expect, it } from "vitest";
import { createTimelineFixtures } from "./timelineFixtures";
import { groupTimelineEvents, timelineDateGroupLabel } from "./groupTimelineEvents";

describe("groupTimelineEvents", () => {
  const now = new Date("2026-07-25T18:00:00.000Z");
  const fixtures = createTimelineFixtures(now);

  it("labels today and yesterday", () => {
    expect(timelineDateGroupLabel(now, now)).toBe("Today");
    const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000);
    expect(timelineDateGroupLabel(yesterday, now)).toBe("Yesterday");
  });

  it("groups all events by day (no filter or range on client)", () => {
    const groups = groupTimelineEvents(fixtures, now);
    expect(groups.map((g) => g.label)).toEqual([
      "Today",
      "Yesterday",
      "2 days ago",
    ]);
    expect(groups.reduce((n, g) => n + g.events.length, 0)).toBe(8);
  });

  it("sorts newest first within each group", () => {
    const groups = groupTimelineEvents(fixtures, now);
    const today = groups.find((g) => g.label === "Today");
    expect(today).toBeDefined();
    // ev-1 at 10:42 should come before ev-2 at 10:31 in newest-first order
    expect(today!.events[0].id).toBe("ev-1");
    expect(today!.events[1].id).toBe("ev-2");
  });

  it("returns empty array for empty input", () => {
    const groups = groupTimelineEvents([], now);
    expect(groups).toHaveLength(0);
  });
});
