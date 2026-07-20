import { useQuery, type UseQueryResult } from "@tanstack/react-query";
import { getProject } from "@/api";
import type { Project } from "@/types";
import { projectQueryKeys } from "@/lib/projectQueryKeys";
import { QUERY_POLICY } from "@/lib/queryPolicy";

export function useProject(
  projectId: string,
  options?: { enabled?: boolean },
): UseQueryResult<Project, Error> {
  const enabled = (options?.enabled ?? true) && Boolean(projectId);
  return useQuery({
    queryKey: projectQueryKeys.detail(projectId),
    queryFn: ({ signal }) => getProject(projectId, { signal }),
    enabled,
    staleTime: QUERY_POLICY.shellStaleTimeMs,
  });
}
