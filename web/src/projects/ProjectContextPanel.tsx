import type { ProjectContextItem } from "@/types";
import { useProjectContext } from "./hooks";
import { ProjectContextImportMemoryModal } from "./ProjectContextImportMemoryModal";
import { ProjectContextPanelWorkspace } from "./ProjectContextPanelWorkspace";
import { firstProjectContextMutationError } from "./projectContextPanelHelpers";
import { useProjectContextMutations } from "./mutations";
import { useProjectContextFormState } from "./useProjectContextFormState";

type Props = {
  projectId: string;
};

const EMPTY_CONTEXT_ITEMS: ProjectContextItem[] = [];

export function ProjectContextPanel({ projectId }: Props) {
  const context = useProjectContext(projectId, { enabled: Boolean(projectId) });
  const mutations = useProjectContextMutations(projectId);
  const form = useProjectContextFormState(mutations);

  const mutationError = firstProjectContextMutationError(mutations);
  const items = context.data?.items ?? EMPTY_CONTEXT_ITEMS;

  return (
    <section className="pc__workspace">
      <ProjectContextImportMemoryModal
        open={form.importOpen}
        onClose={() => form.setImportOpen(false)}
        isPending={mutations.createContextMutation.isPending}
        onImport={form.submitImport}
      />
      {mutationError ? (
        <div className="pd__inline-error" role="alert">
          {mutationError.message}
        </div>
      ) : null}
      <ProjectContextPanelWorkspace
        items={items}
        isLoading={context.isLoading}
        error={context.error}
        mutations={mutations}
        onImportMemory={() => form.setImportOpen(true)}
      />
    </section>
  );
}
