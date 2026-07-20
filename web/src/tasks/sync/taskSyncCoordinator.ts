import type { QueryClient } from "@tanstack/react-query";
import { parseTaskChangeFrame } from "../task-query";
import {
  applyFlushDecision,
  applySyncEffects,
  flushProgressStreams,
  mergePendingDelta,
} from "./applySyncEffects";
import { decideFlushBatch } from "./decideFlushBatch";
import { decideSyncFrame } from "./decideSyncFrame";
import {
  clearPending,
  debounceDelayMs,
  emptyPending,
  PROGRESS_STREAM_INVALIDATE_MAX_WAIT_MS,
  PROGRESS_STREAM_INVALIDATE_WINDOW_MS,
  SSE_INVALIDATE_MAX_WAIT_MS,
  SSE_INVALIDATE_WINDOW_MS,
  type PendingInvalidations,
  type PendingProgressStreams,
} from "./syncConstants";
import { shouldSuppressTaskMutationEcho } from "./mutationGuard";
import type { SyncSchedule } from "./syncTypes";

export type FrameDispatchResult =
  | { kind: "debounce" }
  | { kind: "immediate" }
  | { kind: "resync" }
  | { kind: "ignore" };

function scheduleToDispatchResult(schedule: SyncSchedule): FrameDispatchResult {
  if (schedule === "resync") {
    return { kind: "resync" };
  }
  if (schedule === "immediate") {
    return { kind: "immediate" };
  }
  if (schedule === "ignore") {
    return { kind: "ignore" };
  }
  return { kind: "debounce" };
}

export type TaskSyncCoordinator = {
  pending: PendingInvalidations;
  pendingProgressStreams: PendingProgressStreams;
  /** Parse + Decide + Apply frame; schedules debounce/progress timers when needed. */
  handleRawFrame: (data: string) => FrameDispatchResult;
  flushStreamInvalidation: () => void;
  flushProgressStreamInvalidation: () => void;
  /**
   * Drop pending maps and cancel timers without flushing.
   * Unmount / dispose = drop (matches historical hook behaviour).
   */
  dispose: () => void;
};

/**
 * Holds pending invalidation maps **and** debounce timers (ADR-0022).
 * The React hook only connects EventSource and toggles the live flag.
 */
export function createTaskSyncCoordinator(queryClient: QueryClient): TaskSyncCoordinator {
  const pending = emptyPending();
  const pendingProgressStreams: PendingProgressStreams = new Map();

  let streamDebounce: ReturnType<typeof setTimeout> | undefined;
  let progressDebounce: ReturnType<typeof setTimeout> | undefined;
  let firstQueuedAt: number | null = null;
  let firstProgressQueuedAt: number | null = null;
  let active = true;

  function flushStreamInvalidation() {
    firstQueuedAt = null;
    const flushDecision = decideFlushBatch(pending);
    clearPending(pending);
    applyFlushDecision(queryClient, flushDecision.invalidateKeys);
  }

  function flushProgressStreamInvalidation() {
    firstProgressQueuedAt = null;
    flushProgressStreams(queryClient, pendingProgressStreams);
  }

  function scheduleDebouncedFlush() {
    const now = Date.now();
    if (firstQueuedAt === null) {
      firstQueuedAt = now;
    }
    const delay = debounceDelayMs(
      now,
      firstQueuedAt,
      SSE_INVALIDATE_WINDOW_MS,
      SSE_INVALIDATE_MAX_WAIT_MS,
    );
    if (streamDebounce !== undefined) {
      clearTimeout(streamDebounce);
    }
    streamDebounce = setTimeout(() => {
      streamDebounce = undefined;
      if (!active) {
        return;
      }
      flushStreamInvalidation();
    }, delay);
  }

  function scheduleProgressStreamInvalidation(taskId: string, cycleId: string) {
    const streamKey = `${taskId}\u0000${cycleId}`;
    pendingProgressStreams.set(streamKey, { taskId, cycleId });
    const now = Date.now();
    if (firstProgressQueuedAt === null) {
      firstProgressQueuedAt = now;
    }
    const delay = debounceDelayMs(
      now,
      firstProgressQueuedAt,
      PROGRESS_STREAM_INVALIDATE_WINDOW_MS,
      PROGRESS_STREAM_INVALIDATE_MAX_WAIT_MS,
    );
    if (progressDebounce !== undefined) {
      clearTimeout(progressDebounce);
    }
    progressDebounce = setTimeout(() => {
      progressDebounce = undefined;
      if (!active) {
        return;
      }
      flushProgressStreamInvalidation();
    }, delay);
  }

  function clearTimers() {
    if (streamDebounce !== undefined) {
      clearTimeout(streamDebounce);
      streamDebounce = undefined;
    }
    if (progressDebounce !== undefined) {
      clearTimeout(progressDebounce);
      progressDebounce = undefined;
    }
    firstQueuedAt = null;
    firstProgressQueuedAt = null;
  }

  return {
    pending,
    pendingProgressStreams,
    handleRawFrame(data) {
      const frame = parseTaskChangeFrame(data);
      const decision = decideSyncFrame({
        frame,
        shouldSuppressTaskEcho: shouldSuppressTaskMutationEcho,
      });

      const enrichmentMarks = applySyncEffects(queryClient, decision.effects);
      mergePendingDelta(pending, decision.pendingDelta);
      mergePendingDelta(pending, enrichmentMarks);

      for (const effect of decision.effects) {
        if (effect.kind === "queue_progress_stream") {
          scheduleProgressStreamInvalidation(effect.taskId, effect.cycleId);
        }
      }

      const result = scheduleToDispatchResult(decision.schedule);
      if (result.kind === "immediate" || result.kind === "ignore") {
        return result;
      }
      if (result.kind === "resync") {
        if (streamDebounce !== undefined) {
          clearTimeout(streamDebounce);
          streamDebounce = undefined;
        }
        firstQueuedAt = null;
        return result;
      }
      scheduleDebouncedFlush();
      return result;
    },
    flushStreamInvalidation,
    flushProgressStreamInvalidation,
    dispose() {
      active = false;
      clearTimers();
      clearPending(pending);
      pendingProgressStreams.clear();
    },
  };
}
