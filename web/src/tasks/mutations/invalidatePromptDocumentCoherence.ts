import type { QueryClient } from "@tanstack/react-query";
import type { TaskInvalidationScope } from "@/lib/queryInvalidation";
import type { PromptSourceKind } from "@/tasks/prompt-editor/types";
import { invalidateTaskCache } from "./invalidateTaskCache";

/**
 * Caches that render a prompt document's name or body, by source kind.
 * The Prompt IDE writes through its document adapter rather than a mutation
 * hook, so this table is what keeps those writes coherent with the lists.
 */
export function promptDocumentCoherenceScopes(
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

/** Refresh every cache that renders the given prompt document. */
export function invalidatePromptDocumentCoherence(
  queryClient: QueryClient,
  kind: PromptSourceKind,
  id: string,
): void {
  const scopes = promptDocumentCoherenceScopes(kind, id);
  if (scopes.length === 0) return;
  invalidateTaskCache(queryClient, ...scopes);
}
