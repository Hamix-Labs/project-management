import type { MutableRefObject } from "react";
import type { FormEvent } from "react";
import {
  addChecklistItem,
  deleteChecklistItem,
  patchChecklistItemText,
  patchChecklistItemVerifyCommands,
} from "@/api";
import { normalizeVerifyCommands } from "@/tasks/task-compose/checklistRequirement";
import {
  beginGuardedTaskWrite,
  endGuardedTaskWrite,
  recordOptimisticApplied,
} from "@/tasks/mutations";
import { taskQueryKeys } from "@/tasks/task-query";
import type {
  ChecklistVerifyCommandInput,
  TaskChecklistItemView,
  TaskChecklistResponse,
} from "@/types";
import type { useChecklistModalControls } from "./useChecklistModalControls";
import {
  type ChecklistMutationDeps,
  type ChecklistOptimisticContext,
  finalizeGuardedChecklistMutationSuccess,
  handleGuardedChecklistMutationError,
  invalidateTaskChecklistQueries,
  nextOptimisticChecklistId,
  snapshotChecklist,
} from "./checklistOptimistic";

type GuardedChecklistMutationConfig<TVars, TData> = {
  deps: ChecklistMutationDeps;
  rumKind: import("@/observability").RUMMutationKind;
  mutationFn: (vars: TVars) => Promise<TData>;
  applyOptimistic: (
    prev: TaskChecklistResponse | undefined,
    vars: TVars,
  ) => { next: TaskChecklistResponse; tempItemId?: string } | null;
  errorToast: string;
  shouldRestore: (context: ChecklistOptimisticContext) => boolean;
  onSuccess?: (
    vars: TVars,
    context: ChecklistOptimisticContext | undefined,
  ) => void | boolean | Promise<void | boolean>;
};

export function buildGuardedChecklistMutation<TVars, TData>(
  config: GuardedChecklistMutationConfig<TVars, TData>,
) {
  const {
    deps,
    rumKind,
    mutationFn,
    applyOptimistic,
    errorToast,
    shouldRestore,
    onSuccess,
  } = config;
  const { taskId, queryClient, optimisticMutationsEnabled } = deps;

  return {
    mutationFn,
    onMutate: async (vars: TVars) => {
      const guard = beginGuardedTaskWrite({
        taskId,
        optimisticEnabled: optimisticMutationsEnabled,
        rumKind,
      });
      if (!guard.guarded) {
        return { prev: undefined, startedAtMs: guard.startedAtMs, guarded: false };
      }
      await queryClient.cancelQueries({
        queryKey: taskQueryKeys.checklist(taskId),
      });
      const prev = snapshotChecklist(queryClient, taskId);
      const patch = applyOptimistic(prev, vars);
      if (patch) {
        queryClient.setQueryData(taskQueryKeys.checklist(taskId), patch.next);
      }
      recordOptimisticApplied(rumKind, guard.startedAtMs);
      return {
        prev,
        startedAtMs: guard.startedAtMs,
        tempItemId: patch?.tempItemId,
        guarded: true,
      };
    },
    onError: (
      _err: unknown,
      _vars: TVars,
      context: ChecklistOptimisticContext | undefined,
    ) => {
      handleGuardedChecklistMutationError(rumKind, context, deps, {
        toastMessage: errorToast,
        shouldRestore,
      });
    },
    onSuccess: async (
      _data: TData,
      variables: TVars,
      context: ChecklistOptimisticContext | undefined,
    ) => {
      await invalidateTaskChecklistQueries(queryClient, taskId);
      let finalize = true;
      if (onSuccess) {
        const result = await onSuccess(variables, context);
        if (result === false) {
          finalize = false;
        }
      }
      if (finalize) {
        finalizeGuardedChecklistMutationSuccess(rumKind, context);
      }
    },
    onSettled: (
      _data: TData | undefined,
      _err: unknown,
      _vars: TVars | undefined,
      context: ChecklistOptimisticContext | undefined,
    ) => {
      if (context?.guarded) {
        endGuardedTaskWrite(taskId);
      }
    },
  };
}

interface AddChecklistMutationDeps extends ChecklistMutationDeps {
  addSubmissionTokenRef: MutableRefObject<number>;
  setNewChecklistText: (text: string) => void;
  setNewChecklistVerifyCommands: (commands: ChecklistVerifyCommandInput[]) => void;
  setChecklistModalOpen: (open: boolean) => void;
}

export function buildAddChecklistMutationOptions(deps: AddChecklistMutationDeps) {
  const { taskId } = deps;
  return buildGuardedChecklistMutation({
    deps,
    rumKind: "checklist_add",
    mutationFn: (input: {
      text: string;
      verify_commands: ChecklistVerifyCommandInput[];
      submissionToken: number;
    }) =>
      addChecklistItem(taskId, input.text, {
        verify_commands: input.verify_commands,
      }),
    applyOptimistic: (prev, input) => {
      const tempId = nextOptimisticChecklistId();
      const sortOrder = prev?.items.length
        ? Math.max(...prev.items.map((i) => i.sort_order)) + 1
        : 0;
      const synthetic: TaskChecklistItemView = {
        id: tempId,
        sort_order: sortOrder,
        text: input.text,
        done: false,
      };
      return {
        next: { items: [...(prev?.items ?? []), synthetic] },
        tempItemId: tempId,
      };
    },
    errorToast: "Couldn't add criterion - reverted.",
    shouldRestore: (ctx) => ctx.tempItemId !== undefined,
    onSuccess: async (variables) => {
      if (deps.addSubmissionTokenRef.current !== variables.submissionToken) {
        return false;
      }
      deps.setNewChecklistText("");
      deps.setNewChecklistVerifyCommands([]);
      deps.setChecklistModalOpen(false);
    },
  });
}

export function buildUpdateChecklistTextMutationOptions(deps: ChecklistMutationDeps) {
  const { taskId } = deps;
  return buildGuardedChecklistMutation({
    deps,
    rumKind: "checklist_edit",
    mutationFn: (input: { itemId: string; text: string }) =>
      patchChecklistItemText(taskId, input.itemId, input.text),
    applyOptimistic: (prev, input) => {
      if (!prev) {
        return null;
      }
      return {
        next: {
          items: prev.items.map((it) =>
            it.id === input.itemId ? { ...it, text: input.text } : it,
          ),
        },
      };
    },
    errorToast: "Couldn't update criterion - reverted.",
    shouldRestore: (ctx) => ctx.prev !== undefined,
  });
}

export function buildUpdateChecklistVerifyCommandsMutationOptions(
  deps: ChecklistMutationDeps,
) {
  const { taskId } = deps;
  return buildGuardedChecklistMutation({
    deps,
    rumKind: "checklist_edit",
    mutationFn: (input: { itemId: string; verify_commands: ChecklistVerifyCommandInput[] }) =>
      patchChecklistItemVerifyCommands(taskId, input.itemId, input.verify_commands),
    applyOptimistic: (prev, input) => {
      if (!prev) {
        return null;
      }
      const cmds = normalizeVerifyCommands(input.verify_commands);
      return {
        next: {
          items: prev.items.map((it) =>
            it.id === input.itemId ? { ...it, verify_commands: cmds } : it,
          ),
        },
      };
    },
    errorToast: "Couldn't update verify commands — reverted.",
    shouldRestore: (ctx) => ctx.prev !== undefined,
  });
}

export function buildDeleteChecklistMutationOptions(deps: ChecklistMutationDeps) {
  const { taskId } = deps;
  return buildGuardedChecklistMutation({
    deps,
    rumKind: "checklist_delete",
    mutationFn: (itemId: string) => deleteChecklistItem(taskId, itemId),
    applyOptimistic: (prev, itemId) => {
      if (!prev) {
        return null;
      }
      return {
        next: {
          items: prev.items.filter((it) => it.id !== itemId),
        },
      };
    },
    errorToast: "Couldn't delete criterion - reverted.",
    shouldRestore: (ctx) => ctx.prev !== undefined,
  });
}

export function createSubmitNewChecklistCriterionHandler(
  modal: ReturnType<typeof useChecklistModalControls>,
  addChecklistMutation: {
    isPending: boolean;
    mutate: (input: {
      text: string;
      verify_commands: ChecklistVerifyCommandInput[];
      submissionToken: number;
    }) => void;
  },
): (e: FormEvent) => void {
  return (e: FormEvent) => {
    e.preventDefault();
    const t = modal.newChecklistText.trim();
    if (!t || addChecklistMutation.isPending) return;
    const submissionToken = ++modal.addSubmissionTokenRef.current;
    const verify_commands = normalizeVerifyCommands(modal.newChecklistVerifyCommands);
    addChecklistMutation.mutate({ text: t, verify_commands, submissionToken });
  };
}

export async function submitEditChecklistCriterionForm(
  e: FormEvent,
  deps: {
    modal: ReturnType<typeof useChecklistModalControls>;
    updateChecklistTextMutation: {
      isPending: boolean;
      mutateAsync: (input: { itemId: string; text: string }) => Promise<TaskChecklistResponse>;
    };
    updateChecklistVerifyCommandsMutation: {
      isPending: boolean;
      mutateAsync: (input: {
        itemId: string;
        verify_commands: ChecklistVerifyCommandInput[];
      }) => Promise<TaskChecklistResponse>;
    };
    closeEditCriterionModal: () => void;
  },
): Promise<void> {
  e.preventDefault();
  const id = deps.modal.editingChecklistItemId;
  if (!id) return;
  const newText = deps.modal.editChecklistText.trim();
  if (!newText) return;
  if (
    deps.updateChecklistTextMutation.isPending ||
    deps.updateChecklistVerifyCommandsMutation.isPending
  ) {
    return;
  }
  const newCommands = normalizeVerifyCommands(deps.modal.editChecklistVerifyCommands);
  const textChanged = newText !== deps.modal.editChecklistOriginalText;
  const commandsChanged =
    JSON.stringify(newCommands) !==
    JSON.stringify(normalizeVerifyCommands(deps.modal.editChecklistOriginalVerifyCommands));
  if (!textChanged && !commandsChanged) {
    deps.closeEditCriterionModal();
    return;
  }
  try {
    if (textChanged) {
      await deps.updateChecklistTextMutation.mutateAsync({
        itemId: id,
        text: newText,
      });
    }
    if (commandsChanged) {
      await deps.updateChecklistVerifyCommandsMutation.mutateAsync({
        itemId: id,
        verify_commands: newCommands,
      });
    }
    if (deps.modal.editingChecklistItemIdRef.current === id) {
      deps.closeEditCriterionModal();
    }
  } catch {
    // mutateAsync rejects when the underlying mutation errors.
  }
}
