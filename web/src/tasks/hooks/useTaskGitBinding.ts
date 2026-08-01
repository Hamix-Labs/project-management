import { useQuery } from "@tanstack/react-query";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { QUERY_POLICY } from "@/lib/queryPolicy";
import { resolveTaskGitBinding } from "../task-git/resolveTaskGitBinding";

export function useTaskGitBinding(worktreeId: string | undefined) {
  const wtId = (worktreeId ?? "").trim();

  return useQuery({
    queryKey: gitQueryKeys.taskBinding(wtId),
    queryFn: ({ signal }) => resolveTaskGitBinding(wtId, { signal }),
    enabled: wtId !== "",
    staleTime: QUERY_POLICY.listStaleTimeMs,
  });
}
