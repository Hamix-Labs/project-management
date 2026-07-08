import { useCallback } from "react";
import { patchTask as patchTaskApi } from "@/api";
import { patchTaskPickupInListCaches } from "@/tasks/mutations";
import { useQueryClient } from "@tanstack/react-query";
import {
  useBulkTaskMutation,
  type BulkTaskFailure,
  type BulkTaskResult,
} from "./useBulkTaskMutation";

/**
 * Concurrency cap for bulk PATCHes. Five simultaneous requests is
 * enough to amortise round-trip latency on a normal connection
 * without thundering-herding the API for a 200-row selection.
 */
export const BULK_SCHEDULE_CONCURRENCY = 5;

export type BulkScheduleFailure = BulkTaskFailure;
export type BulkScheduleResult = BulkTaskResult;

/**
 * Fires N parallel PATCH /tasks/{id} requests with `pickup_not_before`
 * set to a single shared value, with a concurrency cap.
 */
export function useBulkScheduleMutation() {
  const queryClient = useQueryClient();
  const { run: runBulk, reset, isPending, lastResult } = useBulkTaskMutation({
    concurrency: BULK_SCHEDULE_CONCURRENCY,
    failureMessage: "Could not update the schedule.",
  });

  const run = useCallback(
    (taskIds: ReadonlyArray<string>, pickupNotBefore: string | null) =>
      runBulk(
        taskIds,
        (id) => patchTaskApi(id, { pickup_not_before: pickupNotBefore }),
        {
          applyOptimistic: (ids) => {
            for (const id of ids) {
              patchTaskPickupInListCaches(queryClient, id, pickupNotBefore);
            }
          },
        },
      ),
    [queryClient, runBulk],
  );

  return { run, reset, isPending, lastResult } as const;
}
