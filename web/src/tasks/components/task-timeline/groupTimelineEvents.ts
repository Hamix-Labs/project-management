import type {
  TimelineDateGroup,
  TimelineEvent,
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

/**
 * Sort newest first and group by relative day label.
 * Server-side filtering (via `since`) is already applied before these events
 * reach the client.
 */
export function groupTimelineEvents(
  events: TimelineEvent[],
  now: Date = new Date(),
): TimelineDateGroup[] {
  const sorted = events.slice().sort((a, b) => Date.parse(b.at) - Date.parse(a.at));

  const order: string[] = [];
  const map = new Map<string, TimelineEvent[]>();
  for (const event of sorted) {
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
