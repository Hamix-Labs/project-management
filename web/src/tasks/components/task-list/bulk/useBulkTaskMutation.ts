import { useCallback, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { errorMessage } from "@/lib/errorMessage";
import {
  runWithConcurrency,
  type RunResult,
} from "@/lib/runWithConcurrency";
import { rumMutationSettled, rumMutationStarted } from "@/observability";
import { useRolloutFlags } from "@/settings";
import {
  invalidateTaskListAndStats,
  recordOptimisticApplied,
} from "@/tasks/mutations";
import {
  beginBulkTaskMutationGuard,
  endBulkTaskMutationGuard,
} from "@/tasks/sync";

export type BulkTaskFailure = {
  taskId: string;
  message: string;
};

export type BulkTaskResult = {
  attempted: number;
  succeeded: number;
  failed: BulkTaskFailure[];
};

type Options = {
  concurrency: number;
  failureMessage: string;
};

export type BulkTaskRunOptions = {
  /** Optimistic list surgery before concurrent API calls (schedule/delete). */
  applyOptimistic?: (taskIds: ReadonlyArray<string>) => void | Promise<void>;
};

export function useBulkTaskMutation({ concurrency, failureMessage }: Options) {
  const queryClient = useQueryClient();
  const { optimisticMutationsEnabled } = useRolloutFlags();
  const [isPending, setPending] = useState(false);
  const [lastResult, setLastResult] = useState<BulkTaskResult | null>(null);
  const inFlightRef = useRef(0);

  const reset = useCallback(() => {
    setLastResult(null);
  }, []);

  const run = useCallback(
    async (
      taskIds: ReadonlyArray<string>,
      runOne: (taskId: string) => Promise<unknown>,
      runOptions?: BulkTaskRunOptions,
    ): Promise<BulkTaskResult> => {
      if (taskIds.length === 0) {
        const empty: BulkTaskResult = {
          attempted: 0,
          succeeded: 0,
          failed: [],
        };
        setLastResult(empty);
        return empty;
      }
      inFlightRef.current += 1;
      setPending(true);
      const startedAtMs = performance.now();
      rumMutationStarted("task_patch");
      beginBulkTaskMutationGuard(taskIds);
      try {
        if (optimisticMutationsEnabled && runOptions?.applyOptimistic) {
          await runOptions.applyOptimistic(taskIds);
          recordOptimisticApplied("task_patch", startedAtMs);
        }

        const calls = taskIds.map((id) => () => runOne(id));
        const results: RunResult<unknown>[] = await runWithConcurrency(
          calls,
          concurrency,
        );
        const failed: BulkTaskFailure[] = [];
        let succeeded = 0;
        for (let i = 0; i < results.length; i++) {
          const r = results[i];
          if (r.ok) {
            succeeded++;
          } else {
            failed.push({
              taskId: taskIds[i],
              message: errorMessage(r.error, failureMessage),
            });
          }
        }
        const summary: BulkTaskResult = {
          attempted: taskIds.length,
          succeeded,
          failed,
        };
        setLastResult(summary);
        if (succeeded > 0) {
          await invalidateTaskListAndStats(queryClient);
        }
        rumMutationSettled("task_patch", performance.now() - startedAtMs, 200);
        return summary;
      } finally {
        endBulkTaskMutationGuard(taskIds);
        inFlightRef.current -= 1;
        if (inFlightRef.current === 0) setPending(false);
      }
    },
    [concurrency, failureMessage, optimisticMutationsEnabled, queryClient],
  );

  return { run, reset, isPending, lastResult } as const;
}
