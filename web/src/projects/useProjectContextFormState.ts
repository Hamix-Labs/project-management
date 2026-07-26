import { useState } from "react";
import type { ProjectContextMutations } from "./projectContextPanelHelpers";

export function useProjectContextFormState(
  mutations: ProjectContextMutations,
) {
  const [importOpen, setImportOpen] = useState(false);

  function submitImport(input: {
    tag: string;
    title: string;
    description: string;
    body: string;
  }) {
    mutations.createContextMutation.mutate(
      {
        tag: input.tag,
        title: input.title,
        description: input.description,
        body: input.body,
        pinned: false,
      },
      {
        onSuccess: () => {
          setImportOpen(false);
        },
      },
    );
  }

  return {
    importOpen,
    setImportOpen,
    submitImport,
  };
}
