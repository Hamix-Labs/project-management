import type { CycleStatus, TaskCycleDetail } from "@/types";
import { formatDurationSeconds } from "@/observability";

export type AttemptTimelineDisplay = {
  startedParts: ReturnType<typeof formatAttemptStartedParts>;
  durationLabel: string;
  showPhaseBadge: boolean;
  endcapLabel: string | null;
  showEndcap: boolean;
  endcapTime: string | null;
  showStartCap: boolean;
  startCapTime: string | null;
};

export function buildAttemptTimelineDisplay(
  cycle: TaskCycleDetail,
  now: number,
): AttemptTimelineDisplay {
  const showPhaseBadge = cycle.phases.length > 1;
  const endcapLabel = attemptEndcapLabel(cycle.status);
  const showEndcap = endcapLabel !== null && cycle.phases.length > 0;
  const endcapTime = showEndcap ? formatAttemptEndedTime(cycle.ended_at) : null;
  const showStartCap = cycle.phases.length > 0;
  const startCapTime = showStartCap
    ? formatAttemptEndedTime(cycle.started_at)
    : null;

  return {
    startedParts: formatAttemptStartedParts(cycle.started_at),
    durationLabel: formatAttemptDurationMeta(
      cycle.started_at,
      cycle.ended_at,
      cycle.status,
      now,
    ),
    showPhaseBadge,
    endcapLabel,
    showEndcap,
    endcapTime,
    showStartCap,
    startCapTime,
  };
}

export function attemptEndcapLabel(status: CycleStatus): string | null {
  switch (status) {
    case "succeeded":
      return "Attempt completed";
    case "failed":
      return "Attempt failed";
    case "aborted":
      return "Attempt aborted";
    default:
      return null;
  }
}

export function formatAttemptEndedTime(
  endedAt: string | undefined,
): string | null {
  if (!endedAt) return null;
  const d = new Date(endedAt);
  if (!Number.isFinite(d.getTime())) return null;
  return d.toLocaleTimeString(undefined, {
    hour: "numeric",
    minute: "2-digit",
  });
}

export function formatAttemptDurationMeta(
  startedAt: string,
  endedAt: string | undefined,
  status: CycleStatus,
  now: number,
): string {
  const start = Date.parse(startedAt);
  const end = endedAt ? Date.parse(endedAt) : now;
  if (!Number.isFinite(start) || !Number.isFinite(end) || end < start) {
    return "Unknown duration";
  }
  const duration = formatDurationSeconds(Math.round((end - start) / 1000));
  const running = status === "running" || !endedAt;
  return running ? `Running for ${duration}` : `Ran for ${duration}`;
}

export function formatAttemptStartedParts(startedAt: string): {
  date: string;
  time: string;
} {
  const started = new Date(startedAt);
  return {
    date: started.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric",
    }),
    time: started.toLocaleTimeString(undefined, {
      hour: "numeric",
      minute: "2-digit",
    }),
  };
}
