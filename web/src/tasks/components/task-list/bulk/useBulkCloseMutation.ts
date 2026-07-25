import { useCallback } from "react";
import { closeTask as closeTaskApi } from "@/api";
import { useQueryClient } from "@tanstack/react-query";
import { taskQueryKeys } from "@/tasks/task-query";
import type { Task, TaskListResponse } from "@/types";
import {
  useBulkTaskMutation,
  type BulkTaskFailure,
  type BulkTaskResult,
} from "./useBulkTaskMutation";

/** Matches `BULK_SCHEDULE_CONCURRENCY` — same thundering-herd rationale. */
export const BULK_CLOSE_CONCURRENCY = 5;

export type BulkCloseFailure = BulkTaskFailure;
export type BulkCloseResult = BulkTaskResult;

/**
 * Fires N `POST /tasks/{id}/close` calls with a concurrency cap. On
 * completion we invalidate the task query namespace (same as bulk
 * schedule PATCH). Optimistically flips each row's status to `closed`
 * in every cached list query so the row does not visually jump between
 * apply and settle.
 */
export function useBulkCloseMutation() {
  const queryClient = useQueryClient();
  const { run: runBulk, reset, isPending, lastResult } = useBulkTaskMutation({
    concurrency: BULK_CLOSE_CONCURRENCY,
    failureMessage: "Could not close the task.",
  });

  const run = useCallback(
    (taskIds: ReadonlyArray<string>) =>
      runBulk(taskIds, (id) => closeTaskApi(id), {
        applyOptimistic: (ids) => {
          const idSet = new Set(ids);
          const listEntries = queryClient.getQueriesData<TaskListResponse>({
            queryKey: taskQueryKeys.listRoot(),
          });
          for (const [key, data] of listEntries) {
            if (!data) continue;
            let changed = false;
            const tasks: Task[] = data.tasks.map((t) => {
              if (!idSet.has(t.id)) return t;
              changed = true;
              return { ...t, status: "closed" as const };
            });
            if (changed) {
              queryClient.setQueryData<TaskListResponse>(key, {
                ...data,
                tasks,
              });
            }
          }
        },
      }),
    [queryClient, runBulk],
  );

  return { run, reset, isPending, lastResult } as const;
}
