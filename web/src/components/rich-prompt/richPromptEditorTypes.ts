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
   * chooses `/ai` from the slash menu. Plan 3 wires this to the real draft
   * agent; for Plan 1 the default host is a no-op — the inline composer
   * opens either way so operators see the shell.
   */
  onAiTrigger?: (msg: string) => void;
};
