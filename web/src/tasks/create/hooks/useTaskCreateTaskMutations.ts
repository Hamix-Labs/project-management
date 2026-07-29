import { useMutation } from "@tanstack/react-query";
import type { MutableRefObject } from "react";
import {
  createTask as apiCreate,
  instantiateTaskTemplates as apiInstantiateTemplates,
} from "@/api";
import { rumMutationSettled } from "@/observability";
import {
  applyCreatedTaskToCache,
  applyCreatedTasksToCache,
  beginGuardedTaskWrite,
  endGuardedTaskWrite,
  invalidateTaskCacheAsync,
  invalidateTaskListAndStats,
} from "@/tasks/mutations";
import {
  beginBulkTaskMutationGuard,
  endBulkTaskMutationGuard,
} from "@/tasks/sync";
import { normalizeChecklistItems } from "../../task-compose/checklistRequirement";
import type { CreateTaskMutationInput } from "../types";

export function useTaskCreateTaskMutations(input: {
  queryClient: import("@tanstack/react-query").QueryClient;
  newDraftIDRef: MutableRefObject<string>;
  closeCreateModal: () => void;
}) {
  const createMutation = useMutation({
    mutationFn: async (mutationInput: CreateTaskMutationInput) => {
      const task = await apiCreate({
        title: mutationInput.title,
        initial_prompt: mutationInput.initial_prompt,
        status: mutationInput.status,
        priority: mutationInput.priority,
        draft_id: mutationInput.draft_id,
        runner: mutationInput.runner,
        cursor_model: mutationInput.cursor_model,
        ...(mutationInput.project_id ? { project_id: mutationInput.project_id } : {}),
        ...(mutationInput.pickup_not_before !== null
          ? { pickup_not_before: mutationInput.pickup_not_before }
          : {}),
        ...(mutationInput.tags.length > 0 ? { tags: mutationInput.tags } : {}),
        ...(mutationInput.milestone ? { milestone: mutationInput.milestone } : {}),
        ...(mutationInput.depends_on.length > 0
          ? { depends_on: mutationInput.depends_on }
          : {}),
        ...(mutationInput.repository_id
          ? { repository_id: mutationInput.repository_id }
          : {}),
        checklist_items: normalizeChecklistItems(mutationInput.checklistItems),
      });
      return { task, input: mutationInput };
    },
    onSuccess: async (result, variables) => {
      // I3 - close modal only when create succeeded for the active draft.
      if (input.newDraftIDRef.current === variables.draft_id) {
        input.closeCreateModal();
      }
      const guard = beginGuardedTaskWrite({
        taskId: result.task.id,
        // Post-create cache seeding always arms the guard so enriched
        // task_created SSE echoes do not race narrow invalidation.
        optimisticEnabled: true,
        rumKind: "task_create",
      });
      try {
        applyCreatedTaskToCache(input.queryClient, result.task);
        await invalidateTaskListAndStats(input.queryClient);
        await invalidateTaskCacheAsync(input.queryClient, { scope: "drafts" });
        rumMutationSettled("task_create", performance.now() - guard.startedAtMs, 201);
      } finally {
        if (guard.guarded) {
          endGuardedTaskWrite(result.task.id);
        }
      }
    },
  });

  const instantiateTemplatesMutation = useMutation({
    mutationFn: (items: import("@/api").TaskTemplateInstantiateItem[]) =>
      apiInstantiateTemplates(items),
    onSuccess: async (result) => {
      if (result.tasks.length === 0) {
        return;
      }
      const taskIds = result.tasks.map((task) => task.id);
      beginBulkTaskMutationGuard(taskIds);
      const startedAtMs = performance.now();
      try {
        applyCreatedTasksToCache(input.queryClient, result.tasks);
        await invalidateTaskListAndStats(input.queryClient);
        rumMutationSettled("task_create", performance.now() - startedAtMs, 201);
      } finally {
        endBulkTaskMutationGuard(taskIds);
      }
    },
  });

  return { createMutation, instantiateTemplatesMutation };
}
