import { useQuery } from "@tanstack/react-query";
import { fetchRepoFile, repoQueryKeys, type RepoFileResult } from "@/api/repo";

export function useRepoFileContent(path: string, worktreeId?: string) {
  const wt = worktreeId?.trim() ?? "";
  const p = path.trim();
  return useQuery({
    queryKey: repoQueryKeys.file(wt, p),
    queryFn: ({ signal }): Promise<RepoFileResult | null> =>
      fetchRepoFile(p, { signal, worktreeId: wt }),
    enabled: Boolean(wt && p),
    staleTime: 60_000,
    gcTime: 30 * 60_000,
    retry: 1,
  });
}
