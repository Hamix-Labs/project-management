import type { QueryClient } from "@tanstack/react-query";
import type { QueryInvalidationKey } from "./types";

export function applyQueryInvalidations(
  queryClient: QueryClient,
  keys: readonly QueryInvalidationKey[],
): void {
  for (const queryKey of keys) {
    void queryClient.invalidateQueries({ queryKey });
  }
}
