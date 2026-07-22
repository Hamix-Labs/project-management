import type { TaskCycle } from "@/types";
import { formatDurationSeconds } from "@/observability";

/** Earliest cycle start across attempts (wall-clock task start). */
export function earliestCycleStartedAt(
  cycles: ReadonlyArray<Pick<TaskCycle, "started_at">>,
): string | null {
  let earliestMs = Number.POSITIVE_INFINITY;
  let earliestIso: string | null = null;
  for (const cycle of cycles) {
    const iso = cycle.started_at.trim();
    if (iso === "") continue;
    const ms = Date.parse(iso);
    if (!Number.isFinite(ms) || ms >= earliestMs) continue;
    earliestMs = ms;
    earliestIso = iso;
  }
  return earliestIso;
}

/**
 * Wall-clock duration from first cycle start to criteria satisfaction.
 * Returns null when either timestamp is missing or the span is invalid.
 */
export function formatTaskCompletionDuration(
  startedAt: string | null | undefined,
  completedAt: string | null | undefined,
): string | null {
  const startIso = (startedAt ?? "").trim();
  const endIso = (completedAt ?? "").trim();
  if (startIso === "" || endIso === "") return null;
  const start = Date.parse(startIso);
  const end = Date.parse(endIso);
  if (!Number.isFinite(start) || !Number.isFinite(end) || end < start) {
    return null;
  }
  return formatDurationSeconds((end - start) / 1000);
}
