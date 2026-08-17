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
  /**
   * Fired when the user presses Space at the start of an empty block or
   * chooses Ask AI from the slash menu.
   */
  onAiTrigger?: (msg: string) => void;
  /**
   * `"full"` (default) shows the Notion floating format bar on selection.
   * `"none"` hides it (trailing `menuRight` still renders).
   */
  toolbar?: "full" | "none";
  /** Optional trailing meta slot (e.g. word count). */
  menuRight?: ReactNode;
  /** Notifies when the TipTap editor instance is ready (or destroyed). */
  onEditorReady?: (editor: Editor | null) => void;
};
