import type { QueryClient } from "@tanstack/react-query";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import type { Task, TaskListResponse } from "@/types";

export function removeTaskFromList(
  list: TaskListResponse,
  removeId: string,
): TaskListResponse | null {
  const nextTasks = list.tasks.filter((t) => t.id !== removeId);
  if (nextTasks.length === list.tasks.length) return null;
  return { ...list, tasks: nextTasks };
}

export function insertTaskInList(
  list: TaskListResponse,
  task: Task,
): TaskListResponse | null {
  if (list.tasks.some((row) => row.id === task.id)) {
    return null;
  }
  return { ...list, tasks: [task, ...list.tasks] };
}

export function patchTaskPickupInList(
  list: TaskListResponse,
  taskId: string,
  pickupNotBefore: string | null,
): TaskListResponse | null {
  let changed = false;
  const tasks = list.tasks.map((task) => {
    if (task.id !== taskId) return task;
    changed = true;
    return {
      ...task,
      pickup_not_before: pickupNotBefore ?? undefined,
    };
  });
  if (!changed) return null;
  return { ...list, tasks };
}

export function applyCreatedTaskToCache(queryClient: QueryClient, task: Task): void {
  queryClient.setQueryData(taskQueryKeys.detail(task.id), task);
  const listEntries = queryClient.getQueriesData<TaskListResponse>({
    queryKey: taskQueryKeys.listRoot(),
  });
  for (const [key, data] of listEntries) {
    if (!data) continue;
    const next = insertTaskInList(data, task);
    if (next) {
      queryClient.setQueryData<TaskListResponse>(key, next);
    }
  }
}

export function applyCreatedTasksToCache(
  queryClient: QueryClient,
  tasks: readonly Task[],
): void {
  for (const task of tasks) {
    applyCreatedTaskToCache(queryClient, task);
  }
}

export function patchTaskPickupInListCaches(
  queryClient: QueryClient,
  taskId: string,
  pickupNotBefore: string | null,
): void {
  const listEntries = queryClient.getQueriesData<TaskListResponse>({
    queryKey: taskQueryKeys.listRoot(),
  });
  for (const [key, data] of listEntries) {
    if (!data) continue;
    const next = patchTaskPickupInList(data, taskId, pickupNotBefore);
    if (next) {
      queryClient.setQueryData<TaskListResponse>(key, next);
    }
  }
}

export function removeTaskFromListCaches(
  queryClient: QueryClient,
  taskId: string,
): void {
  const listEntries = queryClient.getQueriesData<TaskListResponse>({
    queryKey: taskQueryKeys.listRoot(),
  });
  for (const [key, data] of listEntries) {
    if (!data) continue;
    const next = removeTaskFromList(data, taskId);
    if (next) {
      queryClient.setQueryData<TaskListResponse>(key, next);
    }
  }
}
