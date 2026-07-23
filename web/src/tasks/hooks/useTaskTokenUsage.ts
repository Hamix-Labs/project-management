import { useQuery, type UseQueryResult } from "@tanstack/react-query";
import { getTaskTokenUsage } from "@/api";
import type { TaskTokenUsageResponse } from "@/types";
import { taskQueryKeys } from "../task-query";

/**
 * Fetches task-wide token accounting for the detail toolbar and cycle
 * history share labels. Invalidated alongside cycle list queries when
 * SSE reports cycle changes (see `decideFlushBatch`).
 */
export function useTaskTokenUsage(
  taskId: string,
  options?: { enabled?: boolean },
): UseQueryResult<TaskTokenUsageResponse, Error> {
  const enabled = (options?.enabled ?? true) && Boolean(taskId);
  return useQuery({
    queryKey: taskQueryKeys.tokenUsage(taskId),
    queryFn: ({ signal }) => getTaskTokenUsage(taskId, { signal }),
    enabled,
  });
}
