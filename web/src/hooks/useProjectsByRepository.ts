import { useQuery } from "@tanstack/react-query";
import { listProjectsByRepository } from "@/api/gitGlobal";
import { gitQueryKeys } from "@/lib/gitQueryKeys";

export function useProjectsByRepository(
  repositoryId: string,
  options?: { enabled?: boolean },
) {
  const id = (repositoryId ?? "").trim();
  return useQuery({
    queryKey: gitQueryKeys.projectsByRepo(id),
    queryFn: ({ signal }) => listProjectsByRepository(id, { signal }),
    enabled: options?.enabled !== false && id !== "",
  });
}
