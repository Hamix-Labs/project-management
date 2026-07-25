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

  it("groups all events under 7d by default fixtures", () => {
    const groups = groupTimelineEvents(fixtures, "all", "7d", now);
    expect(groups.map((g) => g.label)).toEqual([
      "Today",
      "Yesterday",
      "2 days ago",
    ]);
    expect(groups.reduce((n, g) => n + g.events.length, 0)).toBe(8);
  });

  it("filters verification category", () => {
    const groups = groupTimelineEvents(fixtures, "verification", "7d", now);
    const kinds = groups.flatMap((g) => g.events.map((e) => e.kind));
    expect(kinds).toEqual(["verification-passed", "verification-failed"]);
  });

  it("excludes older events for 24h range", () => {
    const groups = groupTimelineEvents(fixtures, "all", "24h", now);
    const ids = groups.flatMap((g) => g.events.map((e) => e.id));
    expect(ids).toContain("ev-1");
    expect(ids).not.toContain("ev-8");
  });
});
