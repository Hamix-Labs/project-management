import { listGlobalGitBranches } from "@/api/gitGlobal";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { useRepoScopedQuery } from "@/hooks/useRepoScopedQuery";

export function useGlobalBranches(
  repositoryId: string,
  options?: { enabled?: boolean },
) {
  return useRepoScopedQuery({
    repositoryId,
    queryKey: gitQueryKeys.globalBranches(repositoryId),
    queryFn: ({ signal }) => listGlobalGitBranches(repositoryId, { signal }),
    options,
  });
}
