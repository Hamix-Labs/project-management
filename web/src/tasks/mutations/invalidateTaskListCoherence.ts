import type { QueryClient } from "@tanstack/react-query";
import { invalidateTaskCacheAsync } from "./invalidateTaskCache";

/** Narrow invalidation aligned with decideFlushBatch list/stats keys. */
export async function invalidateTaskListAndStats(
  queryClient: QueryClient,
): Promise<void> {
  await invalidateTaskCacheAsync(queryClient, { scope: "listStats" });
}
