import { listProjectsByRepository } from "@/api/gitGlobal";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { useRepoScopedQuery } from "@/hooks/useRepoScopedQuery";

export function useProjectsByRepository(
  repositoryId: string,
  options?: { enabled?: boolean },
) {
  const id = (repositoryId ?? "").trim();
  return useRepoScopedQuery({
    repositoryId: id,
    queryKey: gitQueryKeys.projectsByRepo(id),
    queryFn: ({ signal }) => listProjectsByRepository(id, { signal }),
    options,
  });
}
