import type { TaskEvent } from "@/types";

/** Cycle id from phase/cycle lifecycle payloads; undefined for other families. */
export function taskEventCycleId(ev: TaskEvent): string | undefined {
  switch (ev.type) {
    case "phase_started":
    case "phase_completed":
    case "phase_failed":
    case "phase_skipped":
    case "cycle_started":
    case "cycle_completed":
    case "cycle_failed":
      return ev.data.cycle_id;
    default:
      return undefined;
  }
}

/** Phase seq from phase lifecycle payloads; undefined otherwise. */
export function taskEventPhaseSeq(ev: TaskEvent): number | undefined {
  switch (ev.type) {
    case "phase_started":
    case "phase_completed":
    case "phase_failed":
    case "phase_skipped":
      return ev.data.phase_seq;
    default:
      return undefined;
  }
}
