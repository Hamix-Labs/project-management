import type { TaskStatsResponse } from "@/types";

/**
 * One-line summary under "All tasks" (reference: "15 shown · 7 ready · 2 in review · 2 blocked").
 * Returns undefined when there is nothing useful to show.
 */
export function formatTaskListHeadingSummary(
  shownCount: number,
  stats: TaskStatsResponse | null | undefined,
): string | undefined {
  if (shownCount === 0 && !stats) return undefined;
  if (shownCount === 0 && stats && stats.total <= 0) return undefined;

  const segments: string[] = [`${shownCount} shown`];

  if (!stats) return segments.join(" · ");

  const ready = stats.ready ?? 0;
  const review = stats.by_status?.review ?? 0;
  const blocked = stats.by_status?.blocked ?? 0;

  if (ready > 0) segments.push(`${ready} ready`);
  if (review > 0) segments.push(`${review} in review`);
  if (blocked > 0) segments.push(`${blocked} blocked`);

  return segments.join(" · ");
}
