import type {
  TaskCycle,
  TaskCyclePhase,
  TaskCyclesListResponse,
} from "@/types/cycle";
import { formatInAppTimezone } from "@/shared/time/appTimezone";
import { formatTaskCompletionDuration } from "../schedule/taskCompletionDuration";

/**
 * Splits the cycle list into the (at most one) running cycle and the
 * full history. The history *includes* the running cycle so the list
 * order stays stable across the running→terminal transition (the row
 * stays in place; only its status pill flips). The running cycle is
 * surfaced separately so the live ticker can render above without
 * having to scan the list. The backend orders cycles newest-first.
 */
export function splitRunningAndHistory(
  envelope: TaskCyclesListResponse | undefined,
): { runningCycle: TaskCycle | null; historyCycles: TaskCycle[] } {
  if (!envelope) return { runningCycle: null, historyCycles: [] };
  const running = envelope.cycles.find((c) => c.status === "running") ?? null;
  return { runningCycle: running, historyCycles: envelope.cycles };
}

export function indexCyclesById(
  cycles: ReadonlyArray<TaskCycle>,
): Map<string, TaskCycle> {
  const map = new Map<string, TaskCycle>();
  for (const cycle of cycles) {
    map.set(cycle.id, cycle);
  }
  return map;
}

export function pickRunningPhase(
  phases: ReadonlyArray<TaskCyclePhase>,
): TaskCyclePhase | null {
  return phases.find((p) => p.status === "running") ?? null;
}

export function pickLatestPhase(
  phases: ReadonlyArray<TaskCyclePhase>,
): TaskCyclePhase | null {
  if (phases.length === 0) return null;
  let best: TaskCyclePhase = phases[0];
  for (const p of phases) {
    if (p.phase_seq > best.phase_seq) best = p;
  }
  return best;
}

export function elapsedSeconds(isoStart: string, now: number): number {
  const start = Date.parse(isoStart);
  if (!Number.isFinite(start)) return 0;
  return Math.max(0, (now - start) / 1000);
}

export type AttemptTiming = {
  /** Formatted pickup timestamp in the operator timezone. */
  startedAt: string;
  /** Formatted completion timestamp, or null while the attempt is running. */
  endedAt: string | null;
  /** Wall-clock duration label (e.g. `1 min`), or null while in progress. */
  duration: string | null;
  /** True when the cycle has no ended_at / is still running. */
  inProgress: boolean;
  /** Accessible full description for the timing region. */
  ariaLabel: string;
};

/**
 * Per-attempt pickup / complete / duration fields for cycle history rows.
 * Uses the operator app timezone so timings match the schedule strip.
 */
export function formatAttemptTiming(cycle: TaskCycle, tz: string): AttemptTiming {
  const startedAt = formatInAppTimezone(cycle.started_at, tz);
  if (cycle.status === "running" || !cycle.ended_at) {
    return {
      startedAt,
      endedAt: null,
      duration: null,
      inProgress: true,
      ariaLabel: `Picked up ${startedAt}, still in progress`,
    };
  }
  const endedAt = formatInAppTimezone(cycle.ended_at, tz);
  const duration = formatTaskCompletionDuration(cycle.started_at, cycle.ended_at);
  const ariaLabel = duration
    ? `Picked up ${startedAt}, completed ${endedAt}, ran ${duration}`
    : `Picked up ${startedAt}, completed ${endedAt}`;
  return {
    startedAt,
    endedAt,
    duration,
    inProgress: false,
    ariaLabel,
  };
}
