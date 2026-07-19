import type { useProjectContextMutations } from "./mutations";

export type ProjectContextMutations = ReturnType<typeof useProjectContextMutations>;

export function firstProjectContextMutationError(
  mutations: ProjectContextMutations,
): Error | null {
  return (
    (mutations.createContextMutation.error as Error | null) ??
    (mutations.patchContextMutation.error as Error | null) ??
    (mutations.deleteContextMutation.error as Error | null)
  );
}
