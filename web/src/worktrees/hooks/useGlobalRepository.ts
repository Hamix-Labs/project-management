import { getGlobalGitRepository } from "@/api/gitGlobal";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { useRepoScopedQuery } from "@/hooks/useRepoScopedQuery";

export function useGlobalRepository(
  repositoryId: string,
  options?: { enabled?: boolean },
) {
  return useRepoScopedQuery({
    repositoryId,
    queryKey: gitQueryKeys.globalRepository(repositoryId),
    queryFn: ({ signal }) => getGlobalGitRepository(repositoryId, { signal }),
    options,
  });
}
