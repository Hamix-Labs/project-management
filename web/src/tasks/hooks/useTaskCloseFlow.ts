import { useCallback, useState } from "react";
import { closeTask } from "../../api";
import { errorMessage } from "@/lib/errorMessage";
import { taskQueryKeys } from "../task-query";
import { cancelQueriesForKeys } from "@/tasks/mutations";
import { useGuardedTaskMutation } from "./useGuardedTaskMutation";
import { useTaskReopenMutation } from "./useTaskReopenMutation";
import {
  applyCloseOptimistic,
  restoreCloseOptimistic,
  type CloseSnapshot,
  type CloseVariables,
} from "./taskCloseOptimistic";
import { taskDisplayRef } from "@/lib/taskShortId";
import type { Task } from "@/types";

/** Subset of `Task` the confirm dialog needs; widened so callers can pass plain rows. */
export type CloseTargetInput = {
  id: string;
  title: string;
  number?: number | null;
};

export type CloseTarget = {
  id: string;
  title: string;
  number?: number | null;
};

export type { CloseVariables };

export type UseTaskCloseFlowResult = {
  closeTarget: CloseTarget | null;
  requestClose: (t: CloseTargetInput) => void;
  cancelClose: () => void;
  confirmClose: () => void;
  closePending: boolean;
  closeError: string | null;
  closeSuccess: boolean;
  closeVariables: CloseVariables | undefined;
  resetCloseError: () => void;
  reopen: (id: string) => void;
  reopenPending: boolean;
  reopenError: string | null;
  resetReopenError: () => void;
};

/**
 * Owns the in-app close-confirmation flow and guarded optimistic close.
 * Hard delete is retired; `POST /tasks/{id}/close` sets terminal
 * `closed`, reversible via `/reopen` (see docs/api.md).
 */
export function useTaskCloseFlow(opts: {
  onClosed?: (id: string) => void;
} = {}): UseTaskCloseFlowResult {
  const { onClosed } = opts;
  const [closeTarget, setCloseTarget] = useState<CloseTarget | null>(null);
  const reopenFlow = useTaskReopenMutation();

  const guarded = useGuardedTaskMutation<CloseVariables, CloseSnapshot, Task>({
    rumKind: "task_close",
    mutationFn: (input) => closeTask(input.id),
    applyOptimistic: async ({ queryClient, variables: input, guard }) => {
      await cancelQueriesForKeys(queryClient, [
        taskQueryKeys.listRoot(),
        taskQueryKeys.detail(input.id),
      ]);
      return applyCloseOptimistic({ queryClient, variables: input, guard });
    },
    restoreOptimistic: restoreCloseOptimistic,
    didRollBack: (context) => !!context.detail || context.lists.length > 0,
    errorToast: "Couldn't close - reverted.",
    onSuccessSideEffect: ({ variables }) => {
      const closedId = variables.id;
      setCloseTarget((prev) => (prev?.id === closedId ? null : prev));
      onClosed?.(closedId);
    },
  });

  const requestClose = useCallback((t: CloseTargetInput) => {
    setCloseTarget({
      id: t.id,
      title: t.title,
      number: t.number ?? null,
    });
  }, []);

  const cancelClose = useCallback(() => {
    setCloseTarget(null);
  }, []);

  const confirmClose = useCallback(() => {
    if (!closeTarget) return;
    guarded.mutate({ id: closeTarget.id });
  }, [closeTarget, guarded]);

  const resetCloseError = useCallback(() => {
    if (!guarded.isError) return;
    guarded.reset();
  }, [guarded]);

  return {
    closeTarget,
    requestClose,
    cancelClose,
    confirmClose,
    closePending: guarded.isPending,
    closeError: guarded.isError ? errorMessage(guarded.error) : null,
    closeSuccess: guarded.isSuccess,
    closeVariables: guarded.variables,
    resetCloseError,
    ...reopenFlow,
  };
}

/** Copy used in the close confirm dialog. */
export function closeConfirmDescription(target: {
  id: string;
  number?: number | null;
}): string {
  return `Stops execution and closes ${taskDisplayRef(target)}. You can reopen later.`;
}
