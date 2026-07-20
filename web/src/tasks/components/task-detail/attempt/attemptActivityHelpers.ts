import { listTaskEvents } from "@/api";
import type { TaskCycleStreamEvent } from "@/types";
import { taskEventCycleId } from "@/tasks/task-events/taskEventFields";

export function sortStreamEventsNewestFirst(
  events: readonly TaskCycleStreamEvent[],
): TaskCycleStreamEvent[] {
  return [...events].sort((a, b) => b.stream_seq - a.stream_seq);
}

export function filterAuditEventsForCycle(
  events: Awaited<ReturnType<typeof listTaskEvents>>["events"] | undefined,
  cycleId: string,
) {
  return (
    events?.filter((ev) => taskEventCycleId(ev) === cycleId) ?? []
  ).sort((a, b) => b.seq - a.seq);
}
