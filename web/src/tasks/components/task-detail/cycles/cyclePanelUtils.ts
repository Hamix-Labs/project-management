import type {
  TaskCycle,
  TaskCyclePhase,
  TaskCyclesListResponse,
} from "@/types/cycle";
import { formatDurationSeconds } from "@/observability";

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

export function formatStartedToEnded(cycle: TaskCycle): string {
  const start = formatLocalTime(cycle.started_at);
  if (cycle.status === "running" || !cycle.ended_at) {
    return `${start} → in progress`;
  }
  const end = formatLocalTime(cycle.ended_at);
  return `${start} → ${end}`;
}

export function formatLocalTime(iso: string): string {
  const ts = Date.parse(iso);
  if (!Number.isFinite(ts)) return iso;
  const d = new Date(ts);
  return d.toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatPhaseDuration(phase: TaskCyclePhase, now: number): string {
  const start = Date.parse(phase.started_at);
  if (!Number.isFinite(start)) return "—";
  const end = phase.ended_at ? Date.parse(phase.ended_at) : now;
  if (!Number.isFinite(end) || end < start) return "—";
  return formatDurationSeconds((end - start) / 1000);
}
