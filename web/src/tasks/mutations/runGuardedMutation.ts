import type { RUMMutationKind } from "@/observability";
import {
  beginGuardedTaskWrite,
  endGuardedTaskWrite,
  recordOptimisticApplied,
  type GuardedWriteContext,
} from "./guardedTaskWrite";

export type RunGuardedTaskMutationOptions<T> = {
  taskId: string;
  optimisticEnabled: boolean;
  rumKind: RUMMutationKind;
  applyOptimistic?: () => void | Promise<void>;
  run: () => Promise<T>;
};

export type RunGuardedTaskMutationResult<T> = {
  value: T;
  guard: GuardedWriteContext;
};

/**
 * Runs a single-task mutation inside the mutation-guard window so SSE
 * echoes are suppressed while cache effects land. Callers that need
 * guard to outlive this function (e.g. useMutation onSettled) should
 * use beginGuardedTaskWrite / endGuardedTaskWrite directly instead.
 */
export async function runGuardedTaskMutation<T>(
  options: RunGuardedTaskMutationOptions<T>,
): Promise<RunGuardedTaskMutationResult<T>> {
  const guard = beginGuardedTaskWrite({
    taskId: options.taskId,
    optimisticEnabled: options.optimisticEnabled,
    rumKind: options.rumKind,
  });
  try {
    if (guard.guarded && options.applyOptimistic) {
      await options.applyOptimistic();
      recordOptimisticApplied(options.rumKind, guard.startedAtMs);
    }
    const value = await options.run();
    return { value, guard };
  } finally {
    if (guard.guarded) {
      endGuardedTaskWrite(options.taskId);
    }
  }
}
