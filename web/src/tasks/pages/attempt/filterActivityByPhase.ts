import type { TaskCycleStreamEvent } from "@/types";
import type { TaskEvent } from "@/types/task";
import { taskEventPhaseSeq } from "@/tasks/task-events/taskEventFields";

export function filterStreamEventsByPhase(
  events: readonly TaskCycleStreamEvent[],
  phaseSeq: number | null,
): TaskCycleStreamEvent[] {
  if (phaseSeq === null) {
    return [...events];
  }
  return events.filter((ev) => ev.phase_seq === phaseSeq);
}

export function filterAuditEventsByPhase(
  events: readonly TaskEvent[],
  phaseSeq: number | null,
): TaskEvent[] {
  if (phaseSeq === null) {
    return [...events];
  }
  return events.filter((ev) => taskEventPhaseSeq(ev) === phaseSeq);
}

export function activityCountCaption(
  filtered: number,
  total: number,
): string | undefined {
  if (filtered === total) {
    return undefined;
  }
  return `${filtered} of ${total}`;
}
