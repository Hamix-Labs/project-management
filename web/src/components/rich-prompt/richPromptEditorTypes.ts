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
};
