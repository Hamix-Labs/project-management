import type { QueryClient } from "@tanstack/react-query";
import type { Task, TaskListResponse } from "@/types";
import { taskQueryKeys } from "@/lib/taskQueryKeys";

/** Optimistically set title on cached task detail + list/board entries. */
export function patchCachedTaskTitle(
  queryClient: QueryClient,
  taskId: string,
  title: string,
): void {
  const detailKey = taskQueryKeys.detail(taskId);
  const detail = queryClient.getQueryData<Task>(detailKey);
  if (detail) {
    queryClient.setQueryData<Task>(detailKey, { ...detail, title });
  }

  const listEntries = queryClient.getQueriesData<TaskListResponse>({
    queryKey: taskQueryKeys.listRoot(),
  });
  for (const [key, data] of listEntries) {
    if (!data) continue;
    let changed = false;
    const tasks = data.tasks.map((t) => {
      if (t.id !== taskId) return t;
      changed = true;
      return { ...t, title };
    });
    if (changed) {
      queryClient.setQueryData<TaskListResponse>(key, { ...data, tasks });
    }
  }
}
