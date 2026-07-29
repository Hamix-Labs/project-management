import { decideProjectInvalidationKeys } from "@/lib/queryInvalidation";
import { settingsQueryKeys } from "@/lib/settingsQueryKeys";
import { taskQueryKeys } from "../task-query";
import type { DecideSyncFrameInput, SyncFrameDecision } from "./syncTypes";

export function decideSyncFrame(input: DecideSyncFrameInput): SyncFrameDecision {
  const { frame, shouldSuppressTaskEcho } = input;
  if (frame === null) {
    // Malformed / unparseable frames must not schedule a debounce that
    // later flushes empty pending → broad taskQueryKeys.all invalidate.
    return { schedule: "ignore", pendingDelta: {}, effects: [] };
  }
  if (frame.kind === "task") {
    if (shouldSuppressTaskEcho(frame.taskId)) {
      return { schedule: "immediate", pendingDelta: {}, effects: [] };
    }
    const pendingDelta: SyncFrameDecision["pendingDelta"] = {
      addTaskId: frame.taskId,
    };
    const effects: SyncFrameDecision["effects"] = [];
    if (frame.data !== undefined) {
      effects.push({
        kind: "patch_task_detail",
        taskId: frame.taskId,
        data: frame.data,
      });
    }
    return { schedule: "debounce", pendingDelta, effects };
  }
  if (frame.kind === "project") {
    return {
      schedule: "immediate",
      pendingDelta: {},
      effects: decideProjectInvalidationKeys({ scope: "list" }).map((queryKey) => ({
        kind: "invalidate",
        queryKey,
      })),
    };
  }
  if (frame.kind === "task_event") {
    const effects: SyncFrameDecision["effects"] = [
      { kind: "invalidate", queryKey: taskQueryKeys.eventsRoot(frame.taskId) },
      {
        kind: "invalidate",
        queryKey: taskQueryKeys.eventDetail(frame.taskId, frame.eventSeq),
      },
    ];
    return { schedule: "immediate", pendingDelta: {}, effects };
  }
  if (frame.kind === "cycle") {
    // Cycle frames own the phase/cycle ledger only. Do not queue the task
    // id — that would invalidate detailRoot/checklist before harness
    // completions publish task_updated (ADR-0022 checklist ownership).
    const pendingDelta: SyncFrameDecision["pendingDelta"] = {
      addCycle: { taskId: frame.taskId, cycleId: frame.cycleId },
    };
    const effects: SyncFrameDecision["effects"] = [];
    if (frame.data !== undefined) {
      effects.push({
        kind: "patch_cycle_detail",
        taskId: frame.taskId,
        cycleId: frame.cycleId,
        data: frame.data,
      });
    }
    return { schedule: "debounce", pendingDelta, effects };
  }
  if (frame.kind === "resync") {
    return {
      schedule: "resync",
      pendingDelta: { clearAllPending: true },
      effects: [
        { kind: "rum_sse_resync" },
        { kind: "invalidate", queryKey: taskQueryKeys.all },
        { kind: "invalidate", queryKey: taskQueryKeys.stats() },
        { kind: "invalidate", queryKey: taskQueryKeys.cycleFailuresRoot() },
        { kind: "invalidate", queryKey: settingsQueryKeys.app() },
      ],
    };
  }
  if (frame.kind === "progress") {
    return {
      schedule: "immediate",
      pendingDelta: {},
      effects: [
        {
          kind: "push_agent_run_progress",
          payload: {
            taskId: frame.taskId,
            cycleId: frame.cycleId,
            phaseSeq: frame.phaseSeq,
            progress: frame.progress,
          },
        },
        {
          kind: "queue_progress_stream",
          taskId: frame.taskId,
          cycleId: frame.cycleId,
        },
      ],
    };
  }
  if (frame.kind === "settings" || frame.kind === "agent_run_cancelled") {
    return {
      schedule: "immediate",
      pendingDelta: {},
      effects: [
        { kind: "invalidate", queryKey: settingsQueryKeys.app() },
        { kind: "invalidate", queryKey: settingsQueryKeys.modelsRoot() },
      ],
    };
  }
  return { schedule: "ignore", pendingDelta: {}, effects: [] };
}
