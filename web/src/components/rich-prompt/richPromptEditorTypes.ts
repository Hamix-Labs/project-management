import type { Editor } from "@tiptap/core";
import type { ReactNode } from "react";

export type RichPromptEditorProps = {
  id: string;
  value: string;
  onChange: (v: string) => void;
  disabled?: boolean;
  placeholder?: string;
  /** When set, @-mention search is scoped to this worktree. */
  worktreeId?: string;
  /**
   * When worktreeId is empty (create allocates later), resolve the repository
   * main worktree for `@` mentions only.
   */
  repositoryId?: string;
  /**
   * When true, the empty-binding hint asks for a repository (create flow)
   * instead of a worktree.
   */
  preferRepositoryHint?: boolean;
  /** Toolbar chrome: `"text"` (default) or `"icons"` for compose page. */
  menuVariant?: "text" | "icons";
  /** Optional trailing toolbar slot (e.g. word count). */
  menuRight?: ReactNode;
  /** Notifies when the TipTap editor instance is ready (or destroyed). */
  onEditorReady?: (editor: Editor | null) => void;
};
