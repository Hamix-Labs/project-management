import { useCallback, useState } from "react";
import { deleteTask } from "../../api";
import { errorMessage } from "@/lib/errorMessage";
import { taskQueryKeys } from "../task-query";
import {
  cancelQueriesForKeys,
  recordOptimisticApplied,
  removeTaskFromList,
} from "@/tasks/mutations";
import {
  useGuardedTaskMutation,
  type GuardedMutationContextBase,
} from "./useGuardedTaskMutation";
import type { Task, TaskListResponse } from "@/types";

/** Subset of `Task` the confirm dialog needs; widened so callers can pass plain rows. */
export type DeleteTargetInput = {
  id: string;
  title: string;
};

export type DeleteTarget = {
  id: string;
  title: string;
};

export type DeleteVariables = { id: string };

export type UseTaskDeleteFlowResult = {
  /** Currently-confirming target, or null when the dialog is closed. */
  deleteTarget: DeleteTarget | null;
  /** Open the confirmation dialog for `t`. */
  requestDelete: (t: DeleteTargetInput) => void;
  /** Close the confirmation dialog without deleting. */
  cancelDelete: () => void;
  /** Fire the delete for the current `deleteTarget`; no-op if none is set. */
  confirmDelete: () => void;
  deletePending: boolean;
  /** User-presentable error message for the most recent delete attempt, or null. */
  deleteError: string | null;
  /** True from the moment the delete settles successfully until `requestDelete` is called again. */
  deleteSuccess: boolean;
  /** The variables of the most recent settled delete (used by the detail page to navigate away). */
  deleteVariables: DeleteVariables | undefined;
  /**
   * Clear the most recent error without firing a new request. Successful
   * delete variables must remain visible for detail-page navigation after
   * the confirm dialog closes.
   */
  resetError: () => void;
};

interface DeleteSnapshot extends GuardedMutationContextBase {
  detail: Task | undefined;
  lists: Array<{ key: readonly unknown[]; data: TaskListResponse }>;
}

/**
 * Owns the in-app delete-confirmation flow and guarded optimistic delete.
 */
export function useTaskDeleteFlow(opts: {
  onDeleted?: (id: string) => void;
} = {}): UseTaskDeleteFlowResult {
  const { onDeleted } = opts;
  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null);

  const guarded = useGuardedTaskMutation<DeleteVariables, DeleteSnapshot>({
    rumKind: "task_delete",
    mutationFn: (input) => deleteTask(input.id),
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
        taskQueryKeys.listRoot(),
        taskQueryKeys.detail(input.id),
      ]);

      const detailKey = taskQueryKeys.detail(input.id);
      const detailPrev = queryClient.getQueryData<Task>(detailKey);
      queryClient.removeQueries({ queryKey: detailKey });

      const listSnapshots: DeleteSnapshot["lists"] = [];
      const listEntries = queryClient.getQueriesData<TaskListResponse>({
        queryKey: taskQueryKeys.listRoot(),
      });
      for (const [key, data] of listEntries) {
        if (!data) continue;
        listSnapshots.push({ key, data });
        const next = removeTaskFromList(data, input.id);
        if (next) {
          queryClient.setQueryData<TaskListResponse>(key, next);
        }
      }

      recordOptimisticApplied("task_delete", guard.startedAtMs);
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
    errorToast: "Couldn't delete - reverted.",
    onSuccessSideEffect: ({ variables }) => {
      const deletedId = variables.id;
      setDeleteTarget((prev) => (prev?.id === deletedId ? null : prev));
      onDeleted?.(deletedId);
    },
  });

  const requestDelete = useCallback((t: DeleteTargetInput) => {
    setDeleteTarget({
      id: t.id,
      title: t.title,
    });
  }, []);

  const cancelDelete = useCallback(() => {
    setDeleteTarget(null);
  }, []);

  const confirmDelete = useCallback(() => {
    if (!deleteTarget) return;
    guarded.mutate({
      id: deleteTarget.id,
    });
  }, [deleteTarget, guarded]);

  const resetError = useCallback(() => {
    if (!guarded.isError) return;
    guarded.reset();
  }, [guarded]);

  return {
    deleteTarget,
    requestDelete,
    cancelDelete,
    confirmDelete,
    deletePending: guarded.isPending,
    deleteError: guarded.isError ? errorMessage(guarded.error) : null,
    deleteSuccess: guarded.isSuccess,
    deleteVariables: guarded.variables,
    resetError,
  };
}
