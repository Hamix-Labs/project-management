import { decideTaskInvalidationKeys } from "@/lib/queryInvalidation";
import { taskQueryKeys } from "../task-query";
import {
  cycleEnrichmentKey,
  type PendingInvalidations,
} from "./syncConstants";
import type { SyncFlushDecision } from "./syncTypes";

/** List + stats keys shared with mutation invalidation catalog (ADR-0080). */
export function syncListStatsInvalidationKeys() {
  return decideTaskInvalidationKeys({ scope: "listStats" });
}

/**
 * Decides which React Query keys to invalidate for a debounced SSE flush.
 *
 * Enrichment coverage (E1): marking a task enriched means only the task *row*
 * cache is current. It must never suppress invalidation of sibling detail keys
 * that were not patched in the same apply step. Checklist is owned by
 * `pending.tasks` (task_updated), not by cycle hints — skip `detailRoot` still
 * requires an explicit `checklist(taskId)` invalidate for each pending task.
 *
 * Cycle hints (E2): cycle invalidation is independent of whether the task row
 * was enriched — do not skip cycles merely because the task id is also pending.
 * Cycle pending does not invalidate checklist (phase ledger ≠ checklist).
 */
export function decideFlushBatch(pending: PendingInvalidations): SyncFlushDecision {
  const taskIds = [...pending.tasks];
  const enrichedTaskIds = new Set(pending.enrichedTasks);
  const cycleEntries = [...pending.cycles.entries()];
  const enrichedCycles = new Set(pending.enrichedCycles);
  const keys: (readonly unknown[])[] = [];

  if (taskIds.length === 0 && cycleEntries.length === 0) {
    keys.push(
      taskQueryKeys.all,
      taskQueryKeys.stats(),
      taskQueryKeys.cycleFailuresRoot(),
      taskQueryKeys.activityRoot(),
    );
    return { invalidateKeys: keys };
  }

  if (taskIds.length > 0) {
    const allTasksEnriched = taskIds.every((id) => enrichedTaskIds.has(id));
    if (!allTasksEnriched) {
      keys.push(taskQueryKeys.detailRoot());
    }
    // Checklist completions publish task_updated (ADR-0022 / ADR-0026).
    // Always invalidate checklist for pending task ids — including when
    // enrichment skipped detailRoot.
    for (const taskId of taskIds) {
      keys.push(taskQueryKeys.checklist(taskId));
    }
  }

  for (const [taskId, cycleSet] of cycleEntries) {
    const allCyclesEnriched = [...cycleSet].every((cycleId) =>
      enrichedCycles.has(cycleEnrichmentKey(taskId, cycleId)),
    );
    if (!allCyclesEnriched) {
      keys.push(taskQueryKeys.cycles(taskId));
      keys.push(taskQueryKeys.tokenUsage(taskId));
    }
  }

  const commitsTaskIds = new Set(taskIds);
  for (const [taskId] of cycleEntries) {
    commitsTaskIds.add(taskId);
  }
  for (const taskId of commitsTaskIds) {
    keys.push(taskQueryKeys.commits(taskId));
  }

  keys.push(...syncListStatsInvalidationKeys(), taskQueryKeys.cycleFailuresRoot(), taskQueryKeys.activityRoot());

  return { invalidateKeys: keys };
}
