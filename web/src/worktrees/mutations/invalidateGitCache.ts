import type { QueryClient } from "@tanstack/react-query";
import {
  applyQueryInvalidations,
  decideGitInvalidationKeys,
  type GitInvalidationScope,
} from "@/lib/queryInvalidation";

export function invalidateGitCache(
  queryClient: QueryClient,
  scope: GitInvalidationScope,
): void {
  applyQueryInvalidations(queryClient, decideGitInvalidationKeys(scope));
}
