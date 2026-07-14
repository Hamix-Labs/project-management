import { listGlobalGitWorktreeCheckoutStatus } from "@/api/gitGlobal";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { useRepoScopedQuery } from "@/hooks/useRepoScopedQuery";

/**
 * Cadence for checkout-status polling on the worktrees detail page. Task SSE
 * does not cover filesystem edits; reconcile is too heavy for live dirty/sync.
 */
export const WORKTREE_CHECKOUT_STATUS_POLL_INTERVAL_MS = 15_000;

type Options = {
  enabled?: boolean;
};

export function useWorktreeCheckoutStatus(repositoryId: string, options?: Options) {
  return useRepoScopedQuery({
    repositoryId,
    queryKey: gitQueryKeys.globalWorktreeCheckoutStatus(repositoryId),
    queryFn: ({ signal }) => listGlobalGitWorktreeCheckoutStatus(repositoryId, { signal }),
    options: {
      enabled: options?.enabled,
      refetchInterval: WORKTREE_CHECKOUT_STATUS_POLL_INTERVAL_MS,
      refetchIntervalInBackground: false,
      placeholderData: (previous) => previous,
    },
  });
}
