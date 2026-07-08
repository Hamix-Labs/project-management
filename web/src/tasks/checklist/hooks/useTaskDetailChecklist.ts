import { useMutation, useQueryClient, type QueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useRef, useState, type FormEvent, type MutableRefObject } from "react";
import {
  addChecklistItem,
  deleteChecklistItem,
  patchChecklistItemText,
  patchChecklistItemVerifyCommands,
} from "@/api";
import {
  normalizeVerifyCommands,
} from "@/tasks/task-compose/checklistRequirement";
import type { ChecklistVerifyCommandInput } from "@/types";
import {
  rumMutationRolledBack,
  rumMutationSettled,
  type RUMMutationKind,
} from "@/observability";
import { useOptionalToast } from "@/shared/toast";
import { useRolloutFlags } from "@/settings";
import {
  beginGuardedTaskWrite,
  endGuardedTaskWrite,
  recordOptimisticApplied,
} from "@/tasks/mutations";
import { taskQueryKeys } from "@/tasks/task-query";
import type { TaskChecklistItemView, TaskChecklistResponse } from "@/types";

interface ChecklistOptimisticContext {
  prev: TaskChecklistResponse | undefined;
  startedAtMs: number;
  guarded: boolean;
  /** Identifier we used for the synthetic add — onSuccess swaps it
   * for the real id in the cache so subsequent edits don't reference
   * the temp id. */
  tempItemId?: string;
}

interface ChecklistMutationDeps {
  taskId: string;
  queryClient: QueryClient;
  toast: ReturnType<typeof useOptionalToast>;
  optimisticMutationsEnabled: boolean;
}

interface AddChecklistMutationDeps extends ChecklistMutationDeps {
  addSubmissionTokenRef: MutableRefObject<number>;
  setNewChecklistText: (text: string) => void;
  setNewChecklistVerifyCommands: (commands: ChecklistVerifyCommandInput[]) => void;
  setChecklistModalOpen: (open: boolean) => void;
}

let optimisticChecklistTempCounter = 0;
function nextOptimisticChecklistId(): string {
  optimisticChecklistTempCounter += 1;
  return `optimistic-${optimisticChecklistTempCounter}`;
}

function snapshotChecklist(
  queryClient: QueryClient,
  taskId: string,
): TaskChecklistResponse | undefined {
  return queryClient.getQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(taskId));
}

function restoreChecklist(
  queryClient: QueryClient,
  taskId: string,
  prev: TaskChecklistResponse | undefined,
): void {
  if (prev !== undefined) {
    queryClient.setQueryData(taskQueryKeys.checklist(taskId), prev);
  } else {
    queryClient.removeQueries({ queryKey: taskQueryKeys.checklist(taskId) });
  }
}

function recordRollback(
  kind: RUMMutationKind,
  startedAtMs: number,
): void {
  rumMutationRolledBack(kind, performance.now() - startedAtMs);
  rumMutationSettled(kind, performance.now() - startedAtMs, 0);
}

async function invalidateTaskChecklistQueries(
  queryClient: QueryClient,
  taskId: string,
): Promise<void> {
  await queryClient.invalidateQueries({
    queryKey: taskQueryKeys.checklist(taskId),
  });
  await queryClient.invalidateQueries({
    queryKey: taskQueryKeys.detail(taskId),
  });
}

function handleGuardedChecklistMutationError(
  rumKind: RUMMutationKind,
  context: ChecklistOptimisticContext | undefined,
  deps: ChecklistMutationDeps,
  options: {
    toastMessage: string;
    shouldRestore: (context: ChecklistOptimisticContext) => boolean;
  },
): void {
  if (!context) {
    return;
  }
  if (context.guarded && options.shouldRestore(context)) {
    restoreChecklist(deps.queryClient, deps.taskId, context.prev);
    recordRollback(rumKind, context.startedAtMs);
  } else {
    rumMutationSettled(
      rumKind,
      performance.now() - context.startedAtMs,
      0,
    );
  }
  deps.toast.error(options.toastMessage);
}

function finalizeGuardedChecklistMutationSuccess(
  rumKind: RUMMutationKind,
  context: ChecklistOptimisticContext | undefined,
): void {
  if (context) {
    rumMutationSettled(
      rumKind,
      performance.now() - context.startedAtMs,
      200,
    );
  }
}

type GuardedChecklistMutationConfig<TVars, TData> = {
  deps: ChecklistMutationDeps;
  rumKind: RUMMutationKind;
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

function buildGuardedChecklistMutation<TVars, TData>(
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

function buildAddChecklistMutationOptions(deps: AddChecklistMutationDeps) {
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

function buildUpdateChecklistTextMutationOptions(deps: ChecklistMutationDeps) {
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

function buildUpdateChecklistVerifyCommandsMutationOptions(deps: ChecklistMutationDeps) {
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

function buildDeleteChecklistMutationOptions(deps: ChecklistMutationDeps) {
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

function resetChecklistAddFormState(
  setNewChecklistText: (text: string) => void,
  setNewChecklistVerifyCommands: (commands: ChecklistVerifyCommandInput[]) => void,
): void {
  setNewChecklistText("");
  setNewChecklistVerifyCommands([]);
}

function resetEditCriterionFormState(
  setEditingChecklistItemId: (id: string | null) => void,
  setEditChecklistText: (text: string) => void,
  setEditChecklistOriginalText: (text: string) => void,
  setEditChecklistVerifyCommands: (commands: ChecklistVerifyCommandInput[]) => void,
  setEditChecklistOriginalVerifyCommands: (commands: ChecklistVerifyCommandInput[]) => void,
): void {
  setEditingChecklistItemId(null);
  setEditChecklistText("");
  setEditChecklistOriginalText("");
  setEditChecklistVerifyCommands([]);
  setEditChecklistOriginalVerifyCommands([]);
}

function useChecklistModalControls(taskId: string) {
  const [checklistModalOpen, setChecklistModalOpen] = useState(false);
  const [newChecklistText, setNewChecklistText] = useState("");
  const [newChecklistVerifyCommands, setNewChecklistVerifyCommands] = useState<
    ChecklistVerifyCommandInput[]
  >([]);
  const [editCriterionModalOpen, setEditCriterionModalOpen] = useState(false);
  const [editingChecklistItemId, setEditingChecklistItemId] = useState<
    string | null
  >(null);
  const [editChecklistText, setEditChecklistText] = useState("");
  const [editChecklistVerifyCommands, setEditChecklistVerifyCommands] = useState<
    ChecklistVerifyCommandInput[]
  >([]);
  const [editChecklistOriginalText, setEditChecklistOriginalText] =
    useState("");
  const [editChecklistOriginalVerifyCommands, setEditChecklistOriginalVerifyCommands] =
    useState<ChecklistVerifyCommandInput[]>([]);
  const addSubmissionTokenRef = useRef(0);
  const editingChecklistItemIdRef = useRef<string | null>(null);

  useEffect(() => {
    editingChecklistItemIdRef.current = editingChecklistItemId;
  }, [editingChecklistItemId]);

  useEffect(() => {
    setChecklistModalOpen(false);
    resetChecklistAddFormState(setNewChecklistText, setNewChecklistVerifyCommands);
    setEditCriterionModalOpen(false);
    resetEditCriterionFormState(
      setEditingChecklistItemId,
      setEditChecklistText,
      setEditChecklistOriginalText,
      setEditChecklistVerifyCommands,
      setEditChecklistOriginalVerifyCommands,
    );
    addSubmissionTokenRef.current += 1;
  }, [taskId]);

  const closeChecklistModal = useCallback(() => {
    addSubmissionTokenRef.current += 1;
    setChecklistModalOpen(false);
    resetChecklistAddFormState(setNewChecklistText, setNewChecklistVerifyCommands);
  }, []);

  const closeEditCriterionModal = useCallback(() => {
    setEditCriterionModalOpen(false);
    resetEditCriterionFormState(
      setEditingChecklistItemId,
      setEditChecklistText,
      setEditChecklistOriginalText,
      setEditChecklistVerifyCommands,
      setEditChecklistOriginalVerifyCommands,
    );
  }, []);

  const openChecklistModal = useCallback(() => {
    addSubmissionTokenRef.current += 1;
    resetChecklistAddFormState(setNewChecklistText, setNewChecklistVerifyCommands);
    setChecklistModalOpen(true);
    setEditCriterionModalOpen(false);
    resetEditCriterionFormState(
      setEditingChecklistItemId,
      setEditChecklistText,
      setEditChecklistOriginalText,
      setEditChecklistVerifyCommands,
      setEditChecklistOriginalVerifyCommands,
    );
  }, []);

  const openEditCriterionModal = useCallback(
    (itemId: string, text: string, verifyCommands: ChecklistVerifyCommandInput[] = []) => {
      addSubmissionTokenRef.current += 1;
      setEditingChecklistItemId(itemId);
      setEditChecklistText(text);
      setEditChecklistOriginalText(text);
      const cmds = verifyCommands ?? [];
      setEditChecklistVerifyCommands(cmds);
      setEditChecklistOriginalVerifyCommands(cmds);
      setEditCriterionModalOpen(true);
      setChecklistModalOpen(false);
      resetChecklistAddFormState(setNewChecklistText, setNewChecklistVerifyCommands);
    },
    [],
  );

  return {
    checklistModalOpen,
    setChecklistModalOpen,
    newChecklistText,
    setNewChecklistText,
    newChecklistVerifyCommands,
    setNewChecklistVerifyCommands,
    editCriterionModalOpen,
    editingChecklistItemId,
    editChecklistText,
    setEditChecklistText,
    editChecklistVerifyCommands,
    setEditChecklistVerifyCommands,
    editChecklistOriginalText,
    editChecklistOriginalVerifyCommands,
    addSubmissionTokenRef,
    editingChecklistItemIdRef,
    closeChecklistModal,
    closeEditCriterionModal,
    openChecklistModal,
    openEditCriterionModal,
  };
}

function createSubmitNewChecklistCriterionHandler(
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

async function submitEditChecklistCriterionForm(
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
    // mutateAsync rejects when the underlying mutation errors. The
    // mutation's own `error` is now populated and the modal stays
    // open so `MutationErrorBanner` can surface it.
  }
}

export function useTaskDetailChecklist(taskId: string) {
  const queryClient = useQueryClient();
  const toast = useOptionalToast();
  const { optimisticMutationsEnabled } = useRolloutFlags();
  const modal = useChecklistModalControls(taskId);

  const mutationDeps: ChecklistMutationDeps = {
    taskId,
    queryClient,
    toast,
    optimisticMutationsEnabled,
  };

  const addChecklistMutation = useMutation<
    void,
    unknown,
    { text: string; verify_commands: ChecklistVerifyCommandInput[]; submissionToken: number },
    ChecklistOptimisticContext
  >(
    buildAddChecklistMutationOptions({
      ...mutationDeps,
      addSubmissionTokenRef: modal.addSubmissionTokenRef,
      setNewChecklistText: modal.setNewChecklistText,
      setNewChecklistVerifyCommands: modal.setNewChecklistVerifyCommands,
      setChecklistModalOpen: modal.setChecklistModalOpen,
    }),
  );

  const submitNewChecklistCriterion = useCallback(
    createSubmitNewChecklistCriterionHandler(modal, addChecklistMutation),
    [modal, addChecklistMutation],
  );

  const updateChecklistTextMutation = useMutation<
    TaskChecklistResponse,
    unknown,
    { itemId: string; text: string },
    ChecklistOptimisticContext
  >(buildUpdateChecklistTextMutationOptions(mutationDeps));

  const updateChecklistVerifyCommandsMutation = useMutation<
    TaskChecklistResponse,
    unknown,
    { itemId: string; verify_commands: ChecklistVerifyCommandInput[] },
    ChecklistOptimisticContext
  >(buildUpdateChecklistVerifyCommandsMutationOptions(mutationDeps));

  const submitEditChecklistCriterion = useCallback(
    (e: FormEvent) =>
      submitEditChecklistCriterionForm(e, {
        modal,
        updateChecklistTextMutation,
        updateChecklistVerifyCommandsMutation,
        closeEditCriterionModal: modal.closeEditCriterionModal,
      }),
    [modal, updateChecklistTextMutation, updateChecklistVerifyCommandsMutation],
  );

  const deleteChecklistMutation = useMutation<
    void,
    unknown,
    string,
    ChecklistOptimisticContext
  >(buildDeleteChecklistMutationOptions(mutationDeps));

  return {
    checklistModalOpen: modal.checklistModalOpen,
    newChecklistText: modal.newChecklistText,
    setNewChecklistText: modal.setNewChecklistText,
    newChecklistVerifyCommands: modal.newChecklistVerifyCommands,
    setNewChecklistVerifyCommands: modal.setNewChecklistVerifyCommands,
    editCriterionModalOpen: modal.editCriterionModalOpen,
    editingChecklistItemId: modal.editingChecklistItemId,
    editChecklistText: modal.editChecklistText,
    setEditChecklistText: modal.setEditChecklistText,
    editChecklistVerifyCommands: modal.editChecklistVerifyCommands,
    setEditChecklistVerifyCommands: modal.setEditChecklistVerifyCommands,
    closeChecklistModal: modal.closeChecklistModal,
    closeEditCriterionModal: modal.closeEditCriterionModal,
    openChecklistModal: modal.openChecklistModal,
    openEditCriterionModal: modal.openEditCriterionModal,
    addChecklistMutation,
    submitNewChecklistCriterion,
    updateChecklistTextMutation,
    updateChecklistVerifyCommandsMutation,
    submitEditChecklistCriterion,
    deleteChecklistMutation,
  };
}
