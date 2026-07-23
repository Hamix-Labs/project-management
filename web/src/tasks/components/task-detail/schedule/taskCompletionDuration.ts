import { formatDurationSeconds } from "@/observability";

/**
 * Wall-clock duration between two ISO timestamps.
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
