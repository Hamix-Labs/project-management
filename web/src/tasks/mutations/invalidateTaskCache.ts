import type { QueryClient } from "@tanstack/react-query";
import {
  applyQueryInvalidations,
  decideTaskInvalidationKeys,
  type TaskInvalidationScope,
} from "@/lib/queryInvalidation";

function collectTaskInvalidationKeys(
  scopes: TaskInvalidationScope[],
): readonly (readonly unknown[])[] {
  const seen = new Set<string>();
  return scopes
    .flatMap((scope) => decideTaskInvalidationKeys(scope))
    .filter((key) => {
      const id = JSON.stringify(key);
      if (seen.has(id)) return false;
      seen.add(id);
      return true;
    });
}

export function invalidateTaskCache(
  queryClient: QueryClient,
  ...scopes: TaskInvalidationScope[]
): void {
  applyQueryInvalidations(queryClient, collectTaskInvalidationKeys(scopes));
}

export async function invalidateTaskCacheAsync(
  queryClient: QueryClient,
  ...scopes: TaskInvalidationScope[]
): Promise<void> {
  const keys = collectTaskInvalidationKeys(scopes);
  await Promise.all(
    keys.map((queryKey) => queryClient.invalidateQueries({ queryKey })),
  );
}
