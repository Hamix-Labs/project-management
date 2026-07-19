import { EmptyState } from "@/shared/EmptyState";
import type { ProjectContextEdge, ProjectContextItem } from "@/types";
import { ProjectContextListView } from "./ProjectContextListView";
import { ProjectContextTreeView } from "./ProjectContextTreeView";
import type { ContextView, ProjectContextMutations } from "./projectContextPanelHelpers";

type Props = {
  contextView: ContextView;
  onContextViewChange: (view: ContextView) => void;
  items: ProjectContextItem[];
  edges: ProjectContextEdge[];
  isLoading: boolean;
  error: Error | null;
  mutations: ProjectContextMutations;
  onImportMemory: () => void;
  onAddEdge: (sourceId?: string) => void;
};

export function ProjectContextPanelWorkspace({
  contextView,
  onContextViewChange,
  items,
  edges,
  isLoading,
  error,
  mutations,
  onImportMemory,
  onAddEdge,
}: Props) {
  if (isLoading) {
    return (
      <div className="pc__skeleton" aria-hidden="true">
        <div className="pd__shimmer pd__shimmer--card" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="pd__inline-error" role="alert">
        {error.message}
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <EmptyState
        title="No memory files yet"
        description="Import a .txt or .md file to create a memory node agents can use on tasks."
        action={{
          label: "Import memory file",
          onClick: onImportMemory,
        }}
        density="compact"
        hideIcon
      />
    );
  }

  return (
    <>
      <div className="pc__action-bar">
        <div className="pc__actions-left">
          <button
            type="button"
            className="pc__btn-primary"
            onClick={onImportMemory}
          >
            Import memory file
          </button>
          {items.length >= 2 ? (
            <button
              type="button"
              className="pc__btn-secondary"
              onClick={() => onAddEdge()}
            >
              Add connection
            </button>
          ) : null}
        </div>
        <div className="pc__view-toggle" role="tablist" aria-label="Context view">
          <button
            type="button"
            role="tab"
            aria-selected={contextView === "list"}
            onClick={() => onContextViewChange("list")}
          >
            List
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={contextView === "tree"}
            onClick={() => onContextViewChange("tree")}
          >
            Tree
          </button>
        </div>
      </div>
      {contextView === "list" ? (
        <ProjectContextListView
          items={items}
          nodeSaving={mutations.patchContextMutation.isPending}
          nodeDeleting={mutations.deleteContextMutation.isPending}
          onSaveNode={(id, patch) =>
            mutations.patchContextMutation.mutate({ id, ...patch })
          }
          onDeleteNode={(id) => mutations.deleteContextMutation.mutate(id)}
          onAddConnection={onAddEdge}
        />
      ) : (
        <ProjectContextTreeView items={items} edges={edges} />
      )}
    </>
  );
}
