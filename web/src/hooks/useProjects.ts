import { useQuery, type UseQueryResult } from "@tanstack/react-query";
import { listProjects } from "@/api";
import type { ProjectListResponse } from "@/types";
import { projectQueryKeys } from "@/lib/projectQueryKeys";

export function useProjects(options?: {
  includeArchived?: boolean;
  limit?: number;
  /**
   * When false, the query stays mounted (so the cache survives) but
   * does not trigger a network fetch — useful for shell-level callers
   * that want the data lazily, only after a user opens a modal or a
   * project picker.
   */
  enabled?: boolean;
}): UseQueryResult<ProjectListResponse, Error> {
  const includeArchived = options?.includeArchived ?? false;
  const limit = options?.limit ?? 50;
  const enabled = options?.enabled ?? true;
  return useQuery({
    queryKey: projectQueryKeys.list(includeArchived, limit),
    queryFn: ({ signal }) =>
      listProjects({
        signal,
        includeArchived,
        limit,
      }),
    enabled,
  });
}
