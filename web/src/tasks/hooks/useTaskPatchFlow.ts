import { useCallback } from "react";
import { patchTask as patchTaskApi } from "../../api";
import { errorMessage } from "@/lib/errorMessage";
import { taskQueryKeys } from "../task-query";
import type { Task, TaskListResponse } from "@/types";
import {
  cancelQueriesForKeys,
  mergePatchIntoTask,
  patchTaskInList,
  recordOptimisticApplied,
  type PatchMutationInput,
} from "@/tasks/mutations";
import {
  useGuardedTaskMutation,
  type GuardedMutationContextBase,
} from "./useGuardedTaskMutation";

export type TaskPatchInput = PatchMutationInput;

export type UseTaskPatchFlowResult = {
  /** Fire the underlying PATCH /tasks/{id}; surface a banner via `patchError`. */
  patchTask: (input: TaskPatchInput) => void;
  patchPending: boolean;
  /** User-presentable error from the most recent patch attempt, or null. */
  patchError: string | null;
  /**
   * Clear the most recent settled state (error or success) without firing a
   * new request. Lets `useTasksApp` wipe a stale `patchError` when the edit
   * form closes.
   */
  resetError: () => void;
};

interface PatchSnapshot extends GuardedMutationContextBase {
  detail: Task | undefined;
  lists: Array<{ key: readonly unknown[]; data: TaskListResponse }>;
}

/**
 * Owns the "save edits to a task" mutation (optimistic + guarded).
 */
export function useTaskPatchFlow(opts: {
  onPatched?: (id: string) => void;
} = {}): UseTaskPatchFlowResult {
  const { onPatched } = opts;

  const guarded = useGuardedTaskMutation<TaskPatchInput, PatchSnapshot>({
    rumKind: "task_patch",
    mutationFn: (input) =>
      patchTaskApi(input.id, {
        title: input.title,
        initial_prompt: input.initial_prompt,
        status: input.status,
        priority: input.priority,
        project_id: input.project_id,
        project_context_item_ids: input.project_context_item_ids,
        tags: input.tags,
        milestone: input.milestone,
        cursor_model: input.cursor_model,
        ...(input.pickup_not_before !== undefined
          ? { pickup_not_before: input.pickup_not_before }
          : {}),
      }),
    applyOptimistic: async ({ queryClient, variables: input, guard }) => {
      if (!guard.guarded) {
        return {
          detail: undefined,
          lists: [],
          startedAtMs: guard.startedAtMs,
          guarded: false,
        };
      }

      await cancelQueriesForKeys(queryClient, [
        taskQueryKeys.detail(input.id),
        taskQueryKeys.listRoot(),
      ]);

      const detailKey = taskQueryKeys.detail(input.id);
      const detailPrev = queryClient.getQueryData<Task>(detailKey);
      const { id: taskId, ...patchFields } = input;
      if (detailPrev) {
        queryClient.setQueryData<Task>(
          detailKey,
          mergePatchIntoTask(detailPrev, patchFields),
        );
      }

      const listSnapshots: PatchSnapshot["lists"] = [];
      const listEntries = queryClient.getQueriesData<TaskListResponse>({
        queryKey: taskQueryKeys.listRoot(),
      });
      for (const [key, data] of listEntries) {
        if (!data) continue;
        listSnapshots.push({ key, data });
        const next = patchTaskInList(data, taskId, patchFields);
        if (next) {
          queryClient.setQueryData<TaskListResponse>(key, next);
        }
      }

      recordOptimisticApplied("task_patch", guard.startedAtMs);

      return {
        detail: detailPrev,
        lists: listSnapshots,
        startedAtMs: guard.startedAtMs,
        guarded: true,
      };
    },
    restoreOptimistic: ({ queryClient, variables: input, context }) => {
      if (context.detail) {
        queryClient.setQueryData(taskQueryKeys.detail(input.id), context.detail);
      }
      for (const snap of context.lists) {
        queryClient.setQueryData(snap.key, snap.data);
      }
    },
    didRollBack: (context) =>
      !!context.detail || context.lists.length > 0,
    errorToast: "Couldn't save - reverted.",
    onSuccessSideEffect: ({ variables }) => {
      onPatched?.(variables.id);
    },
  });

  const resetError = useCallback(() => {
    if (guarded.isIdle) return;
    guarded.reset();
  }, [guarded]);

  return {
    patchTask: guarded.mutate,
    patchPending: guarded.isPending,
    patchError: guarded.isError ? errorMessage(guarded.error) : null,
    resetError,
  };
}
