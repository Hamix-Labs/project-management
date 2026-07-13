import type { QueryClient } from "@tanstack/react-query";
import {
  rumMutationRolledBack,
  rumMutationSettled,
  type RUMMutationKind,
} from "@/observability";
import { taskQueryKeys } from "@/tasks/task-query";
import type { TaskChecklistResponse } from "@/types";
import { invalidateTaskCacheAsync } from "@/tasks/mutations/invalidateTaskCache";

export interface ChecklistOptimisticContext {
  prev: TaskChecklistResponse | undefined;
  startedAtMs: number;
  guarded: boolean;
  tempItemId?: string;
}

export interface ChecklistMutationDeps {
  taskId: string;
  queryClient: QueryClient;
  toast: { error: (message: string) => void };
  optimisticMutationsEnabled: boolean;
}

let optimisticChecklistTempCounter = 0;

export function nextOptimisticChecklistId(): string {
  optimisticChecklistTempCounter += 1;
  return `optimistic-${optimisticChecklistTempCounter}`;
}

export function snapshotChecklist(
  queryClient: QueryClient,
  taskId: string,
): TaskChecklistResponse | undefined {
  return queryClient.getQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(taskId));
}

export function restoreChecklist(
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

export function recordRollback(kind: RUMMutationKind, startedAtMs: number): void {
  rumMutationRolledBack(kind, performance.now() - startedAtMs);
  rumMutationSettled(kind, performance.now() - startedAtMs, 0);
}

export async function invalidateTaskChecklistQueries(
  queryClient: QueryClient,
  taskId: string,
): Promise<void> {
  await invalidateTaskCacheAsync(queryClient, { scope: "checklist", taskId });
}

export function handleGuardedChecklistMutationError(
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
    rumMutationSettled(rumKind, performance.now() - context.startedAtMs, 0);
  }
  deps.toast.error(options.toastMessage);
}

export function finalizeGuardedChecklistMutationSuccess(
  rumKind: RUMMutationKind,
  context: ChecklistOptimisticContext | undefined,
): void {
  if (context) {
    rumMutationSettled(rumKind, performance.now() - context.startedAtMs, 200);
  }
}
