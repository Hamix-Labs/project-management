import { useQuery, type UseQueryResult } from "@tanstack/react-query";
import { fetchRepoCommitDiff, repoQueryKeys, type RepoDiffResult } from "@/api/repo";

export function useCommitDiff(
  sha: string,
  options?: { worktreeId?: string; enabled?: boolean },
): UseQueryResult<RepoDiffResult | null, Error> {
  const worktreeId = (options?.worktreeId ?? "").trim();
  const enabled =
    (options?.enabled ?? true) && Boolean(sha) && worktreeId !== "";
  return useQuery({
    queryKey: repoQueryKeys.diff(worktreeId, sha),
    queryFn: ({ signal }) =>
      fetchRepoCommitDiff(sha, { worktreeId, signal }),
    enabled,
    staleTime: Number.POSITIVE_INFINITY,
    gcTime: 30 * 60_000,
    retry: 1,
  });
}
