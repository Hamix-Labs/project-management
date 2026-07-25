import type { TimelineRangeId } from "./timelineTypes";

const DAY_MS = 24 * 60 * 60 * 1000;

export type TimelineRangeOption = {
  id: TimelineRangeId;
  label: string;
  /** Lookback window; `null` means all time. */
  ms: number | null;
};

export const TIMELINE_RANGE_OPTIONS: readonly TimelineRangeOption[] = [
  { id: "24h", label: "Last 24 hours", ms: DAY_MS },
  { id: "7d", label: "Last 7 days", ms: 7 * DAY_MS },
  { id: "30d", label: "Last 30 days", ms: 30 * DAY_MS },
  { id: "90d", label: "Last 90 days", ms: 90 * DAY_MS },
  { id: "all", label: "All time", ms: null },
] as const;

export const DEFAULT_TIMELINE_RANGE: TimelineRangeId = "7d";

export function timelineRangeLabel(id: TimelineRangeId): string {
  return (
    TIMELINE_RANGE_OPTIONS.find((o) => o.id === id)?.label ?? "Last 7 days"
  );
}

/** Inclusive lower bound for events; `null` when range is all time. */
export function timelineRangeCutoff(
  rangeId: TimelineRangeId,
  now: Date = new Date(),
): Date | null {
  const opt = TIMELINE_RANGE_OPTIONS.find((o) => o.id === rangeId);
  if (!opt || opt.ms == null) return null;
  return new Date(now.getTime() - opt.ms);
}

export function eventInTimelineRange(
  atIso: string,
  rangeId: TimelineRangeId,
  now: Date = new Date(),
): boolean {
  const cutoff = timelineRangeCutoff(rangeId, now);
  if (!cutoff) return true;
  const at = Date.parse(atIso);
  if (Number.isNaN(at)) return false;
  return at >= cutoff.getTime() && at <= now.getTime();
}
