import { listGlobalGitBranches, listGlobalGitLiveBranches } from "@/api/gitGlobal";
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

export function useGlobalLiveBranches(
  repositoryId: string,
  options?: { enabled?: boolean },
) {
  return useRepoScopedQuery({
    repositoryId,
    queryKey: gitQueryKeys.globalLiveBranches(repositoryId),
    queryFn: ({ signal }) => listGlobalGitLiveBranches(repositoryId, { signal }),
    options,
  });
}
