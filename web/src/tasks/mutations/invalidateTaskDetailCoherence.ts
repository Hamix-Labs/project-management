import type { QueryClient } from "@tanstack/react-query";
import { invalidateTaskCacheAsync } from "./invalidateTaskCache";

export async function invalidateTaskDetailCoherence(
  queryClient: QueryClient,
  taskId: string,
): Promise<void> {
  await invalidateTaskCacheAsync(queryClient, { scope: "detail", taskId });
}
