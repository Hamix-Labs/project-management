import type { QueryClient } from "@tanstack/react-query";
import type { Task, TaskListResponse } from "@/types";
import { taskQueryKeys } from "@/lib/taskQueryKeys";

/**
 * Finds a task already in the React Query cache (detail or any list/board
 * under listRoot). Used as detail `placeholderData` so list→detail navigation
 * can paint immediately without waiting on GET /tasks/{id}.
 */
export function findCachedTask(
  queryClient: QueryClient,
  taskId: string,
): Task | undefined {
  const id = taskId.trim();
  if (id === "") {
    return undefined;
  }

  const detail = queryClient.getQueryData<Task>(taskQueryKeys.detail(id));
  if (detail?.id === id) {
    return detail;
  }

  const listEntries = queryClient.getQueriesData<TaskListResponse>({
    queryKey: taskQueryKeys.listRoot(),
  });
  for (const [, data] of listEntries) {
    if (!data?.tasks) continue;
    const hit = data.tasks.find((task) => task.id === id);
    if (hit) return hit;
  }
  return undefined;
}
