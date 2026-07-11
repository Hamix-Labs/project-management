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
  onAddNode: () => void;
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
  onAddNode,
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
        title="No context nodes yet"
        description="Add memory nodes and connect them as the work evolves."
        action={{
          label: "Add memory",
          onClick: onAddNode,
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
          <button type="button" className="pc__btn-primary" onClick={onAddNode}>
            Add memory
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
