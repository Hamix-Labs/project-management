import { listGlobalGitLiveWorktrees } from "@/api/gitGlobal";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { useRepoScopedQuery } from "@/hooks/useRepoScopedQuery";

export function useGlobalLiveWorktrees(
  repositoryId: string,
  options?: { enabled?: boolean },
) {
  return useRepoScopedQuery({
    repositoryId,
    queryKey: gitQueryKeys.globalLiveWorktrees(repositoryId),
    queryFn: ({ signal }) => listGlobalGitLiveWorktrees(repositoryId, { signal }),
    options,
  });
}
