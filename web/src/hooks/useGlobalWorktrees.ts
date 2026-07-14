import { listGlobalGitWorktrees } from "@/api/gitGlobal";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { useRepoScopedQuery } from "@/hooks/useRepoScopedQuery";

export function useGlobalWorktrees(
  repositoryId: string,
  options?: { enabled?: boolean },
) {
  return useRepoScopedQuery({
    repositoryId,
    queryKey: gitQueryKeys.globalWorktrees(repositoryId),
    queryFn: ({ signal }) => listGlobalGitWorktrees(repositoryId, { signal }),
    options,
  });
}
