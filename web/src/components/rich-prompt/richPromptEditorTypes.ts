export type RichPromptEditorProps = {
  id: string;
  value: string;
  onChange: (v: string) => void;
  disabled?: boolean;
  placeholder?: string;
  /** When set, @-mention search is scoped to this worktree. */
  worktreeId?: string;
};
