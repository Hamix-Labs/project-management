import { useQuery, type UseQueryResult } from "@tanstack/react-query";
import { listProjectContext } from "@/api";
import type { ProjectContextListResponse } from "@/types";
import { projectQueryKeys } from "@/lib/projectQueryKeys";

export function useProjectContext(
  projectId: string,
  options?: { enabled?: boolean; limit?: number; pinnedOnly?: boolean },
): UseQueryResult<ProjectContextListResponse, Error> {
  const enabled = (options?.enabled ?? true) && Boolean(projectId);
  return useQuery({
    queryKey: projectQueryKeys.context(projectId),
    queryFn: ({ signal }) =>
      listProjectContext(projectId, {
        signal,
        limit: options?.limit,
        pinnedOnly: options?.pinnedOnly,
      }),
    enabled,
  });
}
