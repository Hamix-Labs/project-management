import type { QueryClient } from "@tanstack/react-query";
import {
  applyQueryInvalidations,
  decideProjectInvalidationKeys,
  type ProjectInvalidationScope,
} from "@/lib/queryInvalidation";

export function invalidateProjectCache(
  queryClient: QueryClient,
  ...scopes: ProjectInvalidationScope[]
): void {
  const seen = new Set<string>();
  const keys = scopes
    .flatMap((scope) => decideProjectInvalidationKeys(scope))
    .filter((key) => {
      const id = JSON.stringify(key);
      if (seen.has(id)) return false;
      seen.add(id);
      return true;
    });
  applyQueryInvalidations(queryClient, keys);
}
