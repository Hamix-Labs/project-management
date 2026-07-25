import { eventInTimelineRange } from "./timelineRange";
import type {
  TimelineDateGroup,
  TimelineEvent,
  TimelineFilterId,
  TimelineRangeId,
} from "./timelineTypes";

const DAY_MS = 24 * 60 * 60 * 1000;

function startOfLocalDay(d: Date): Date {
  const x = new Date(d.getTime());
  x.setHours(0, 0, 0, 0);
  return x;
}

/** Calendar-day distance from `at` back to `now` (0 = today). */
export function calendarDaysAgo(at: Date, now: Date): number {
  const a = startOfLocalDay(at).getTime();
  const b = startOfLocalDay(now).getTime();
  return Math.max(0, Math.round((b - a) / DAY_MS));
}

export function timelineDateGroupLabel(at: Date, now: Date): string {
  const days = calendarDaysAgo(at, now);
  if (days === 0) return "Today";
  if (days === 1) return "Yesterday";
  return `${days} days ago`;
}

export function formatTimelineClock(at: Date): string {
  return at.toLocaleTimeString(undefined, {
    hour: "numeric",
    minute: "2-digit",
  });
}

function matchesFilter(
  event: TimelineEvent,
  filter: TimelineFilterId,
): boolean {
  if (filter === "all") return true;
  return event.category === filter;
}

/**
 * Filter by category + range, newest first, group by relative day label.
 */
export function groupTimelineEvents(
  events: TimelineEvent[],
  filter: TimelineFilterId,
  rangeId: TimelineRangeId,
  now: Date = new Date(),
): TimelineDateGroup[] {
  const filtered = events
    .filter(
      (e) =>
        matchesFilter(e, filter) && eventInTimelineRange(e.at, rangeId, now),
    )
    .slice()
    .sort((a, b) => Date.parse(b.at) - Date.parse(a.at));

  const order: string[] = [];
  const map = new Map<string, TimelineEvent[]>();
  for (const event of filtered) {
    const at = new Date(event.at);
    const label = timelineDateGroupLabel(at, now);
    if (!map.has(label)) {
      map.set(label, []);
      order.push(label);
    }
    map.get(label)!.push(event);
  }
  return order.map((label) => ({ label, events: map.get(label)! }));
}
