import { useState } from "react";
import type { ProjectContextMutations } from "./projectContextPanelHelpers";

export function useProjectContextFormState(
  mutations: ProjectContextMutations,
) {
  const [importOpen, setImportOpen] = useState(false);

  function submitImport(input: { title: string; body: string }) {
    mutations.createContextMutation.mutate(
      {
        kind: "note",
        title: input.title,
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
