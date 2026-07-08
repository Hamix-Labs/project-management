import type { QueryClient } from "@tanstack/react-query";
import { taskQueryKeys } from "@/lib/taskQueryKeys";

export async function invalidateTaskDetailCoherence(
  queryClient: QueryClient,
  taskId: string,
): Promise<void> {
  await queryClient.invalidateQueries({ queryKey: taskQueryKeys.detail(taskId) });
}
