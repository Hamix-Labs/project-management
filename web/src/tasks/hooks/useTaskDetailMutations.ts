import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { patchTask, retryTask } from "@/api";
import {
  rumMutationRolledBack,
  rumMutationSettled,
} from "@/observability";
import { useOptionalToast } from "@/shared/toast";
import { useRolloutFlags } from "@/settings";
import {
  beginGuardedTaskWrite,
  endGuardedTaskWrite,
  invalidateTaskCacheAsync,
  recordOptimisticApplied,
} from "@/tasks/mutations";
import type { Task } from "@/types";
import type { TaskRetryMode } from "../components/dialogs/TaskRetryConfirmDialog";
import { taskQueryKeys } from "../task-query";

function useTaskDetailRetryMutation(
  taskId: string,
  optimisticMutationsEnabled: boolean,
  onRetryConfirmed: () => void,
) {
  const queryClient = useQueryClient();

  return useMutation<
    unknown,
    unknown,
    TaskRetryMode,
    { prev: Task | undefined; startedAtMs: number; guarded: boolean }
  >({
    mutationFn: (mode) => retryTask(taskId, { mode }),
    onMutate: async () => {
      const guard = beginGuardedTaskWrite({
        taskId,
        optimisticEnabled: optimisticMutationsEnabled,
        rumKind: "task_retry",
      });
      if (!guard.guarded) {
        return { prev: undefined, startedAtMs: guard.startedAtMs, guarded: false };
      }
      await queryClient.cancelQueries({ queryKey: taskQueryKeys.detail(taskId) });
      const detailKey = taskQueryKeys.detail(taskId);
      const prev = queryClient.getQueryData<Task>(detailKey);
      if (prev) {
        queryClient.setQueryData<Task>(detailKey, { ...prev, status: "ready" });
      }
      recordOptimisticApplied("task_retry", guard.startedAtMs);
      return { prev, startedAtMs: guard.startedAtMs, guarded: true };
    },
    onError: (_err, _vars, context) => {
      if (context?.prev) {
        queryClient.setQueryData(taskQueryKeys.detail(taskId), context.prev);
      }
      if (context) {
        if (context.prev !== undefined) {
          rumMutationRolledBack(
            "task_retry",
            performance.now() - context.startedAtMs,
          );
        }
        rumMutationSettled(
          "task_retry",
          performance.now() - context.startedAtMs,
          0,
        );
      }
    },
    onSuccess: async (_data, _vars, context) => {
      onRetryConfirmed();
      await invalidateTaskCacheAsync(queryClient, { scope: "listStats" });
      if (context) {
        rumMutationSettled(
          "task_retry",
          performance.now() - context.startedAtMs,
          200,
        );
      }
    },
    onSettled: (_data, _err, _vars, context) => {
      if (context?.guarded) {
        endGuardedTaskWrite(taskId);
      }
    },
  });
}

function useTaskDetailAutonomyMutation(
  taskId: string,
  optimisticMutationsEnabled: boolean,
  toast: ReturnType<typeof useOptionalToast>,
  onAutonomyConfirmed: () => void,
) {
  const queryClient = useQueryClient();

  return useMutation<
    unknown,
    unknown,
    "ready" | "on_hold",
    { prev: Task | undefined; startedAtMs: number; next: "ready" | "on_hold"; guarded: boolean }
  >({
    mutationFn: (next) => patchTask(taskId, { status: next }),
    onMutate: async (next) => {
      const guard = beginGuardedTaskWrite({
        taskId,
        optimisticEnabled: optimisticMutationsEnabled,
        rumKind: "task_autonomy",
      });
      if (!guard.guarded) {
        return { prev: undefined, startedAtMs: guard.startedAtMs, next, guarded: false };
      }
      await queryClient.cancelQueries({ queryKey: taskQueryKeys.detail(taskId) });
      const detailKey = taskQueryKeys.detail(taskId);
      const prev = queryClient.getQueryData<Task>(detailKey);
      if (prev) {
        queryClient.setQueryData<Task>(detailKey, { ...prev, status: next });
      }
      recordOptimisticApplied("task_autonomy", guard.startedAtMs);
      return { prev, startedAtMs: guard.startedAtMs, next, guarded: true };
    },
    onError: (_err, _vars, context) => {
      if (context?.prev) {
        queryClient.setQueryData(taskQueryKeys.detail(taskId), context.prev);
      }
      if (context) {
        if (context.prev !== undefined) {
          rumMutationRolledBack(
            "task_autonomy",
            performance.now() - context.startedAtMs,
          );
        }
        rumMutationSettled(
          "task_autonomy",
          performance.now() - context.startedAtMs,
          0,
        );
      }
      toast.error("Couldn't update autonomy — reverted.");
    },
    onSuccess: async (_data, _vars, context) => {
      await invalidateTaskCacheAsync(queryClient, { scope: "listStats" });
      onAutonomyConfirmed();
      if (context) {
        rumMutationSettled(
          "task_autonomy",
          performance.now() - context.startedAtMs,
          200,
        );
      }
    },
    onSettled: (_data, _err, _vars, context) => {
      if (context?.guarded) {
        endGuardedTaskWrite(taskId);
      }
    },
  });
}

export function useTaskDetailMutations(taskId: string) {
  const [modelConfigOpen, setModelConfigOpen] = useState(false);
  const [autonomyConfirmOpen, setAutonomyConfirmOpen] = useState(false);
  const [retryConfirmMode, setRetryConfirmMode] = useState<TaskRetryMode | null>(
    null,
  );
  const toast = useOptionalToast();
  const { optimisticMutationsEnabled } = useRolloutFlags();
  const retryMutation = useTaskDetailRetryMutation(
    taskId,
    optimisticMutationsEnabled,
    () => setRetryConfirmMode(null),
  );
  const autonomyMutation = useTaskDetailAutonomyMutation(
    taskId,
    optimisticMutationsEnabled,
    toast,
    () => setAutonomyConfirmOpen(false),
  );

  return {
    modelConfigOpen,
    setModelConfigOpen,
    autonomyConfirmOpen,
    setAutonomyConfirmOpen,
    retryConfirmMode,
    setRetryConfirmMode,
    retryMutation,
    autonomyMutation,
  };
}
