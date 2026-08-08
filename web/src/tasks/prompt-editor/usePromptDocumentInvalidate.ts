import { useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { TaskInvalidationScope } from "@/lib/queryInvalidation";
import { invalidateTaskCache } from "@/tasks/mutations";
import { isPromptSourceKind } from "./promptDocumentAdapter";
import type { PromptSourceKind } from "./types";

/** Caches that render an open document's name or body, per source kind. */
export function promptDocumentInvalidationScopes(
  kind: PromptSourceKind,
  id: string,
): TaskInvalidationScope[] {
  switch (kind) {
    case "draft":
      return [{ scope: "drafts" }];
    case "template":
      return [{ scope: "templates" }];
    case "task":
      return id
        ? [{ scope: "detail", taskId: id }, { scope: "listStats" }]
        : [];
    case "ephemeral":
      return [];
  }
}

/**
 * Refresh the collection that owns the open document after a successful write.
 * Fire-and-forget: mounted lists refetch in the background and unmounted ones
 * are marked stale, so a rename shows up without a reload and leaving the
 * editor never waits on a refetch.
 */
export function usePromptDocumentInvalidate(
  sourceKind: string,
  sourceId: string,
): () => void {
  const queryClient = useQueryClient();
  return useCallback(() => {
    if (!isPromptSourceKind(sourceKind)) return;
    const scopes = promptDocumentInvalidationScopes(sourceKind, sourceId);
    if (scopes.length === 0) return;
    invalidateTaskCache(queryClient, ...scopes);
  }, [queryClient, sourceId, sourceKind]);
}
