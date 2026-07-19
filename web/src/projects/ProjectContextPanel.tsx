import { useMemo } from "react";
import type { ProjectContextEdge, ProjectContextItem } from "@/types";
import { useProjectContext } from "./hooks";
import { ProjectContextAddEdgeModal } from "./ProjectContextAddEdgeModal";
import { ProjectContextImportMemoryModal } from "./ProjectContextImportMemoryModal";
import { ProjectContextPanelWorkspace } from "./ProjectContextPanelWorkspace";
import {
  buildMemorySelectOptions,
  buildRelationSelectOptions,
  buildStrengthSelectOptions,
  firstProjectContextMutationError,
} from "./projectContextPanelHelpers";
import { useProjectContextMutations } from "./mutations";
import { useProjectContextFormState } from "./useProjectContextFormState";

type Props = {
  projectId: string;
};

const EMPTY_CONTEXT_ITEMS: ProjectContextItem[] = [];
const EMPTY_CONTEXT_EDGES: ProjectContextEdge[] = [];

export function ProjectContextPanel({ projectId }: Props) {
  const context = useProjectContext(projectId, { enabled: Boolean(projectId) });
  const mutations = useProjectContextMutations(projectId);
  const form = useProjectContextFormState(mutations);

  const mutationError = firstProjectContextMutationError(mutations);
  const items = context.data?.items ?? EMPTY_CONTEXT_ITEMS;
  const edges = context.data?.edges ?? EMPTY_CONTEXT_EDGES;
  const memoryOptions = useMemo(() => buildMemorySelectOptions(items), [items]);
  const relationOptions = useMemo(() => buildRelationSelectOptions(), []);
  const strengthOptions = useMemo(() => buildStrengthSelectOptions(), []);

  return (
    <section className="pc__workspace">
      <ProjectContextImportMemoryModal
        open={form.importOpen}
        onClose={() => form.setImportOpen(false)}
        isPending={mutations.createContextMutation.isPending}
        onImport={form.submitImport}
      />
      {items.length < 2 ? (
        <p className="pc__hint">
          Add at least two memory nodes to start connecting them.
        </p>
      ) : null}
      <ProjectContextAddEdgeModal
        open={form.addEdgeOpen}
        onClose={() => form.setAddEdgeOpen(false)}
        isPending={mutations.createEdgeMutation.isPending}
        memoryOptions={memoryOptions}
        relationOptions={relationOptions}
        strengthOptions={strengthOptions}
        newEdgeSourceID={form.newEdgeSourceID}
        newEdgeTargetID={form.newEdgeTargetID}
        newEdgeRelation={form.newEdgeRelation}
        newEdgeStrength={form.newEdgeStrength}
        newEdgeNote={form.newEdgeNote}
        newEdgeEditorKey={form.newEdgeEditorKey}
        onSourceChange={form.setNewEdgeSourceID}
        onTargetChange={form.setNewEdgeTargetID}
        onRelationChange={form.setNewEdgeRelation}
        onStrengthChange={form.setNewEdgeStrength}
        onNoteChange={form.setNewEdgeNote}
        onSubmit={form.submitEdge}
      />
      {mutationError ? (
        <div className="pd__inline-error" role="alert">
          {mutationError.message}
        </div>
      ) : null}
      <ProjectContextPanelWorkspace
        contextView={form.contextView}
        onContextViewChange={form.setContextView}
        items={items}
        edges={edges}
        isLoading={context.isLoading}
        error={context.error}
        mutations={mutations}
        onImportMemory={() => form.setImportOpen(true)}
        onAddEdge={form.openAddEdge}
      />
    </section>
  );
}
