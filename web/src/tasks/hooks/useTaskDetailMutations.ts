import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { approveTask, patchTask, polishTask, retryTask } from "@/api";
import type { ChecklistItemDraft } from "@/types";
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
      await invalidateTaskCacheAsync(
        queryClient,
        { scope: "listStats" },
        { scope: "detail", taskId },
      );
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

function useTaskDetailApproveMutation(
  taskId: string,
  optimisticMutationsEnabled: boolean,
  onApproveConfirmed: () => void,
) {
  const queryClient = useQueryClient();

  return useMutation<
    unknown,
    unknown,
    void,
    { prev: Task | undefined; startedAtMs: number; guarded: boolean }
  >({
    mutationFn: () => approveTask(taskId),
    onMutate: async () => {
      const guard = beginGuardedTaskWrite({
        taskId,
        optimisticEnabled: optimisticMutationsEnabled,
        rumKind: "task_approve",
      });
      if (!guard.guarded) {
        return { prev: undefined, startedAtMs: guard.startedAtMs, guarded: false };
      }
      await queryClient.cancelQueries({ queryKey: taskQueryKeys.detail(taskId) });
      const detailKey = taskQueryKeys.detail(taskId);
      const prev = queryClient.getQueryData<Task>(detailKey);
      if (prev) {
        queryClient.setQueryData<Task>(detailKey, { ...prev, status: "done" });
      }
      recordOptimisticApplied("task_approve", guard.startedAtMs);
      return { prev, startedAtMs: guard.startedAtMs, guarded: true };
    },
    onError: (_err, _vars, context) => {
      if (context?.prev) {
        queryClient.setQueryData(taskQueryKeys.detail(taskId), context.prev);
      }
      if (context) {
        if (context.prev !== undefined) {
          rumMutationRolledBack(
            "task_approve",
            performance.now() - context.startedAtMs,
          );
        }
        rumMutationSettled(
          "task_approve",
          performance.now() - context.startedAtMs,
          0,
        );
      }
    },
    onSuccess: async (_data, _vars, context) => {
      onApproveConfirmed();
      await invalidateTaskCacheAsync(
        queryClient,
        { scope: "listStats" },
        { scope: "detail", taskId },
        { scope: "events", taskId },
      );
      if (context) {
        rumMutationSettled(
          "task_approve",
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

function useTaskDetailPolishMutation(
  taskId: string,
  optimisticMutationsEnabled: boolean,
  onPolishConfirmed: () => void,
) {
  const queryClient = useQueryClient();

  return useMutation<
    unknown,
    unknown,
    {
      instructions: string;
      flaggedCriterionIds: string[];
      newCriteria: ChecklistItemDraft[];
    },
    { prev: Task | undefined; startedAtMs: number; guarded: boolean }
  >({
    mutationFn: (input) =>
      polishTask(taskId, {
        instructions: input.instructions,
        flagged_criterion_ids: input.flaggedCriterionIds,
        new_criteria: input.newCriteria.map((item) => ({
          text: item.text,
          verify_commands: item.verify_commands,
        })),
      }),
    onMutate: async () => {
      const guard = beginGuardedTaskWrite({
        taskId,
        optimisticEnabled: optimisticMutationsEnabled,
        rumKind: "task_polish",
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
      recordOptimisticApplied("task_polish", guard.startedAtMs);
      return { prev, startedAtMs: guard.startedAtMs, guarded: true };
    },
    onError: (_err, _vars, context) => {
      if (context?.prev) {
        queryClient.setQueryData(taskQueryKeys.detail(taskId), context.prev);
      }
      if (context) {
        if (context.prev !== undefined) {
          rumMutationRolledBack(
            "task_polish",
            performance.now() - context.startedAtMs,
          );
        }
        rumMutationSettled(
          "task_polish",
          performance.now() - context.startedAtMs,
          0,
        );
      }
    },
    onSuccess: async (_data, _vars, context) => {
      onPolishConfirmed();
      await invalidateTaskCacheAsync(
        queryClient,
        { scope: "listStats" },
        { scope: "detail", taskId },
        { scope: "events", taskId },
        { scope: "checklist", taskId },
      );
      if (context) {
        rumMutationSettled(
          "task_polish",
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
      await invalidateTaskCacheAsync(
        queryClient,
        { scope: "listStats" },
        { scope: "detail", taskId },
      );
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
  const [approveConfirmOpen, setApproveConfirmOpen] = useState(false);
  const [polishDialogOpen, setPolishDialogOpen] = useState(false);
  const toast = useOptionalToast();
  const { optimisticMutationsEnabled } = useRolloutFlags();
  const retryMutation = useTaskDetailRetryMutation(
    taskId,
    optimisticMutationsEnabled,
    () => setRetryConfirmMode(null),
  );
  const approveMutation = useTaskDetailApproveMutation(
    taskId,
    optimisticMutationsEnabled,
    () => setApproveConfirmOpen(false),
  );
  const polishMutation = useTaskDetailPolishMutation(
    taskId,
    optimisticMutationsEnabled,
    () => setPolishDialogOpen(false),
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
    approveConfirmOpen,
    setApproveConfirmOpen,
    polishDialogOpen,
    setPolishDialogOpen,
    retryMutation,
    approveMutation,
    polishMutation,
    autonomyMutation,
  };
}
