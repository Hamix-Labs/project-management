import type { QueryClient } from "@tanstack/react-query";
import { taskQueryKeys } from "@/lib/taskQueryKeys";

/** Narrow invalidation aligned with decideFlushBatch list/stats keys. */
export async function invalidateTaskListAndStats(
  queryClient: QueryClient,
): Promise<void> {
  await queryClient.invalidateQueries({ queryKey: taskQueryKeys.listRoot() });
  await queryClient.invalidateQueries({ queryKey: taskQueryKeys.stats() });
}
