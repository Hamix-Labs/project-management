import { taskQueryKeys } from "../task-query";
import { recordOptimisticApplied } from "@/tasks/mutations";
import type { GuardedMutationContextBase } from "./useGuardedTaskMutation";
import type { QueryClient } from "@tanstack/react-query";
import type { Task, TaskListResponse } from "@/types";

export type CloseVariables = { id: string };

export interface CloseSnapshot extends GuardedMutationContextBase {
  detail: Task | undefined;
  lists: Array<{ key: readonly unknown[]; data: TaskListResponse }>;
}

export async function applyCloseOptimistic(args: {
  queryClient: QueryClient;
  variables: CloseVariables;
  guard: GuardedMutationContextBase;
}): Promise<CloseSnapshot> {
  const { queryClient, variables: input, guard } = args;
  if (!guard.guarded) {
    return {
      detail: undefined,
      lists: [],
      startedAtMs: guard.startedAtMs,
      guarded: false,
    };
  }

  const detailKey = taskQueryKeys.detail(input.id);
  const detailPrev = queryClient.getQueryData<Task>(detailKey);
  if (detailPrev) {
    queryClient.setQueryData<Task>(detailKey, {
      ...detailPrev,
      status: "closed",
    });
  }

  const listSnapshots: CloseSnapshot["lists"] = [];
  const listEntries = queryClient.getQueriesData<TaskListResponse>({
    queryKey: taskQueryKeys.listRoot(),
  });
  for (const [key, data] of listEntries) {
    if (!data) continue;
    listSnapshots.push({ key, data });
    let changed = false;
    const tasks = data.tasks.map((t) => {
      if (t.id !== input.id) return t;
      changed = true;
      return { ...t, status: "closed" as const };
    });
    if (changed) {
      queryClient.setQueryData<TaskListResponse>(key, {
        ...data,
        tasks,
      });
    }
  }

  recordOptimisticApplied("task_close", guard.startedAtMs);
  return {
    detail: detailPrev,
    lists: listSnapshots,
    startedAtMs: guard.startedAtMs,
    guarded: true,
  };
}

export function restoreCloseOptimistic(args: {
  queryClient: QueryClient;
  variables: CloseVariables;
  context: CloseSnapshot;
}): void {
  const { queryClient, variables: input, context } = args;
  if (context.detail) {
    queryClient.setQueryData(taskQueryKeys.detail(input.id), context.detail);
  }
  for (const snap of context.lists) {
    queryClient.setQueryData(snap.key, snap.data);
  }
}
