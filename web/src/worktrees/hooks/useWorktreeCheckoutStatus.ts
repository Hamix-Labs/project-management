import { useQuery } from "@tanstack/react-query";
import { listGlobalGitWorktreeCheckoutStatus } from "@/api/gitGlobal";
import { gitQueryKeys } from "../queryKeys";

/**
 * Cadence for checkout-status polling on the worktrees detail page. Task SSE
 * does not cover filesystem edits; reconcile is too heavy for live dirty/sync.
 */
export const WORKTREE_CHECKOUT_STATUS_POLL_INTERVAL_MS = 15_000;

type Options = {
  enabled?: boolean;
};

export function useWorktreeCheckoutStatus(repositoryId: string, options?: Options) {
  return useQuery({
    queryKey: gitQueryKeys.globalWorktreeCheckoutStatus(repositoryId),
    queryFn: ({ signal }) => listGlobalGitWorktreeCheckoutStatus(repositoryId, { signal }),
    enabled: options?.enabled !== false && repositoryId.trim() !== "",
    refetchInterval: WORKTREE_CHECKOUT_STATUS_POLL_INTERVAL_MS,
    refetchIntervalInBackground: false,
    placeholderData: (previous) => previous,
  });
}
