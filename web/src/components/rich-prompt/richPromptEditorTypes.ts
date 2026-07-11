import type { ProjectContextEdge, ProjectContextItem } from "@/types";

export type RichPromptEditorProjectContextProps = {
  /**
   * All project context items available for the active project. When empty,
   * the `#` suggestion plugin still opens but renders an empty state instead
   * of swallowing the trigger.
   */
  items: ProjectContextItem[];
  /**
   * Project context edges (`source -> target`). Used by the choice dialog to
   * preview how many descendants would be added when the operator picks
   * "Reference this node and its children".
   */
  edges: ProjectContextEdge[];
  /** IDs already on the task. The REFERENCES block renders one row per id. */
  selectedIds: string[];
  /**
   * Replace the selected ids. The editor calls this through the shared
   * `mergeProjectContextSelection`, so callers should not dedupe again.
   */
  onSelectedIdsChange: (ids: string[]) => void;
};

export type RichPromptEditorProps = {
  id: string;
  value: string;
  onChange: (v: string) => void;
  disabled?: boolean;
  placeholder?: string;
  /**
   * When provided, the editor wires the `#` project context suggestion plugin
   * and renders the read-only REFERENCES block above the editable content.
   * Omit on surfaces where project context does not apply (e.g. project edge
   * notes) so behaviour stays unchanged.
   */
  projectContext?: RichPromptEditorProjectContextProps;
  /** When set, @-mention search is scoped to this worktree. */
  worktreeId?: string;
};
