import { useCallback } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { reopenTask } from "../../api";
import { errorMessage } from "@/lib/errorMessage";
import { taskQueryKeys } from "../task-query";
import {
  cancelQueriesForKeys,
  invalidateTaskCacheAsync,
} from "@/tasks/mutations";
import type { Task } from "@/types";

export type UseTaskReopenMutationResult = {
  reopen: (id: string) => void;
  reopenPending: boolean;
  reopenError: string | null;
  resetReopenError: () => void;
};

/**
 * Direct-action reopen (`POST /tasks/{id}/reopen`) with an optimistic
 * flip to `ready`. No confirm dialog — the detail toolbar owns the
 * single-click affordance.
 */
export function useTaskReopenMutation(): UseTaskReopenMutationResult {
  const queryClient = useQueryClient();

  const reopenMutation = useMutation<Task, unknown, string>({
    mutationFn: (id) => reopenTask(id),
    onMutate: async (id) => {
      await cancelQueriesForKeys(queryClient, [
        taskQueryKeys.listRoot(),
        taskQueryKeys.detail(id),
      ]);
      const detailKey = taskQueryKeys.detail(id);
      const prev = queryClient.getQueryData<Task>(detailKey);
      if (prev) {
        queryClient.setQueryData<Task>(detailKey, {
          ...prev,
          status: "ready",
        });
      }
      return { prev } as { prev: Task | undefined };
    },
    onError: (_err, id, context) => {
      const ctx = context as { prev?: Task } | undefined;
      if (ctx?.prev) {
        queryClient.setQueryData(taskQueryKeys.detail(id), ctx.prev);
      }
    },
    onSuccess: async (task) => {
      queryClient.setQueryData(taskQueryKeys.detail(task.id), task);
      await invalidateTaskCacheAsync(
        queryClient,
        { scope: "listStats" },
        { scope: "detail", taskId: task.id },
      );
    },
  });

  const reopen = useCallback(
    (id: string) => {
      reopenMutation.mutate(id);
    },
    [reopenMutation],
  );

  const resetReopenError = useCallback(() => {
    if (!reopenMutation.isError) return;
    reopenMutation.reset();
  }, [reopenMutation]);

  return {
    reopen,
    reopenPending: reopenMutation.isPending,
    reopenError: reopenMutation.isError
      ? errorMessage(reopenMutation.error)
      : null,
    resetReopenError,
  };
}
