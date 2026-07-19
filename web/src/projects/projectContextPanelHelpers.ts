import type { CustomSelectOption } from "@/components/custom-select";
import {
  PROJECT_CONTEXT_RELATIONS,
  type ProjectContextItem,
} from "@/types";
import type { useProjectContextMutations } from "./mutations";

export type ProjectContextMutations = ReturnType<typeof useProjectContextMutations>;

export function firstProjectContextMutationError(
  mutations: ProjectContextMutations,
): Error | null {
  return (
    (mutations.createContextMutation.error as Error | null) ??
    (mutations.patchContextMutation.error as Error | null) ??
    (mutations.deleteContextMutation.error as Error | null) ??
    (mutations.createEdgeMutation.error as Error | null) ??
    (mutations.patchEdgeMutation.error as Error | null) ??
    (mutations.deleteEdgeMutation.error as Error | null)
  );
}

export function buildMemorySelectOptions(items: ProjectContextItem[]): CustomSelectOption[] {
  return [
    { value: "", label: "Select memory" },
    ...items.map((item) => ({ value: item.id, label: item.title })),
  ];
}

export function buildRelationSelectOptions(): CustomSelectOption[] {
  return PROJECT_CONTEXT_RELATIONS.map((relation) => ({
    value: relation,
    label: relation.replace("_", " "),
  }));
}

export function buildStrengthSelectOptions(): CustomSelectOption[] {
  return [1, 2, 3, 4, 5].map((strength) => ({
    value: String(strength),
    label: String(strength),
  }));
}
