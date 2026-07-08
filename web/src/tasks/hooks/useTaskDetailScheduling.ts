import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import {
  addTaskDependency,
  patchTask,
  patchTaskGate,
  removeTaskDependency,
} from "@/api";
import { errorMessage } from "@/lib/errorMessage";
import { useRolloutFlags } from "@/settings";
import type { Task } from "@/types";
import {
  beginGuardedTaskWrite,
  endGuardedTaskWrite,
  invalidateTaskDetailCoherence,
  invalidateTaskListAndStats,
  recordOptimisticApplied,
} from "@/tasks/mutations";
import { taskQueryKeys } from "../task-query";

type GuardedSchedulingContext = {
  prev: Task | undefined;
  startedAtMs: number;
  guarded: boolean;
};

function useGuardedSchedulingHintMutation<TVars>(
  taskId: string,
  optimisticMutationsEnabled: boolean,
  mutationFn: (vars: TVars) => Promise<unknown>,
  onSuccessSideEffect?: (vars: TVars) => void,
) {
  const queryClient = useQueryClient();
  return useMutation<unknown, unknown, TVars, GuardedSchedulingContext>({
    mutationFn,
    onMutate: async () => {
      const guard = beginGuardedTaskWrite({
        taskId,
        optimisticEnabled: optimisticMutationsEnabled,
        rumKind: "task_patch",
      });
      return { prev: undefined, startedAtMs: guard.startedAtMs, guarded: guard.guarded };
    },
    onSuccess: async (_data, vars) => {
      onSuccessSideEffect?.(vars);
      await invalidateTaskDetailCoherence(queryClient, taskId);
    },
    onSettled: (_d, _e, _v, context) => {
      if (context?.guarded) endGuardedTaskWrite(taskId);
    },
  });
}

function useGuardedSchedulingPatchMutation<TVars>(
  taskId: string,
  optimisticMutationsEnabled: boolean,
  config: {
    mutationFn: (vars: TVars) => Promise<unknown>;
    applyOptimistic: (prev: Task, vars: TVars) => Task;
  },
) {
  const queryClient = useQueryClient();
  return useMutation<unknown, unknown, TVars, GuardedSchedulingContext>({
    mutationFn: config.mutationFn,
    onMutate: async (vars) => {
      const guard = beginGuardedTaskWrite({
        taskId,
        optimisticEnabled: optimisticMutationsEnabled,
        rumKind: "task_patch",
      });
      if (!guard.guarded) {
        return { prev: undefined, startedAtMs: guard.startedAtMs, guarded: false };
      }
      await queryClient.cancelQueries({ queryKey: taskQueryKeys.detail(taskId) });
      const detailKey = taskQueryKeys.detail(taskId);
      const prev = queryClient.getQueryData<Task>(detailKey);
      if (prev) queryClient.setQueryData(detailKey, config.applyOptimistic(prev, vars));
      recordOptimisticApplied("task_patch", guard.startedAtMs);
      return { prev, startedAtMs: guard.startedAtMs, guarded: true };
    },
    onError: (_e, _v, context) => {
      if (context?.prev) queryClient.setQueryData(taskQueryKeys.detail(taskId), context.prev);
    },
    onSuccess: async () => {
      await invalidateTaskDetailCoherence(queryClient, taskId);
      await invalidateTaskListAndStats(queryClient);
    },
    onSettled: (_d, _e, _v, context) => {
      if (context?.guarded) endGuardedTaskWrite(taskId);
    },
  });
}

export function useTaskDetailScheduling(taskId: string) {
  const { optimisticMutationsEnabled } = useRolloutFlags();
  const [depAddValue, setDepAddValue] = useState("");
  const [tagsDraft, setTagsDraft] = useState("");
  const [milestoneDraft, setMilestoneDraft] = useState("");

  const addDepMutation = useGuardedSchedulingHintMutation<void>(
    taskId,
    optimisticMutationsEnabled,
    () => addTaskDependency(taskId, depAddValue.trim()),
    () => setDepAddValue(""),
  );
  const removeDepMutation = useGuardedSchedulingHintMutation(
    taskId,
    optimisticMutationsEnabled,
    (dependsOnTaskId: string) => removeTaskDependency(taskId, dependsOnTaskId),
  );
  const gateMutation = useGuardedSchedulingHintMutation(
    taskId,
    optimisticMutationsEnabled,
    (action: "release" | "hold" | "clear_hold") => patchTaskGate(taskId, action),
  );
  const tagsMutation = useGuardedSchedulingPatchMutation(taskId, optimisticMutationsEnabled, {
    mutationFn: (tags: string[]) =>
      patchTask(taskId, { tags: tags.map((t) => t.trim()).filter(Boolean) }),
    applyOptimistic: (prev, tags) => ({
      ...prev,
      tags: tags.map((t) => t.trim()).filter(Boolean),
    }),
  });
  const milestoneMutation = useGuardedSchedulingPatchMutation(
    taskId,
    optimisticMutationsEnabled,
    {
      mutationFn: (milestone: string | null) =>
        patchTask(taskId, { milestone: milestone === "" ? null : milestone }),
      applyOptimistic: (prev, milestone) => ({
        ...prev,
        milestone: milestone === "" ? undefined : milestone ?? undefined,
      }),
    },
  );

  return {
    depAddValue,
    setDepAddValue,
    tagsDraft,
    setTagsDraft,
    milestoneDraft,
    setMilestoneDraft,
    addDepMutation,
    removeDepMutation,
    gateMutation,
    tagsMutation,
    milestoneMutation,
    schedulingError: [
      addDepMutation.error,
      removeDepMutation.error,
      gateMutation.error,
      tagsMutation.error,
      milestoneMutation.error,
    ]
      .map((e) => (e ? errorMessage(e) : null))
      .find(Boolean),
  };
}
