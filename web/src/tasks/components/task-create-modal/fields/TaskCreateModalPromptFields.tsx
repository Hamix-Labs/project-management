import { TaskComposePromptField } from "../../task-compose/fields/TaskComposePromptField";

type Props = {
  idsPrefix: string;
  editorKey: string;
  prompt: string;
  disabled: boolean;
  onPromptChange: (v: string) => void;
  worktreeId?: string;
  repositoryId?: string;
  preferRepositoryHint?: boolean;
};

export function TaskCreateModalPromptFields({
  idsPrefix,
  editorKey,
  prompt,
  disabled,
  onPromptChange,
  worktreeId,
  repositoryId,
  preferRepositoryHint = false,
}: Props) {
  return (
    <TaskComposePromptField
      idsPrefix={idsPrefix}
      editorKey={editorKey}
      prompt={prompt}
      disabled={disabled}
      onPromptChange={onPromptChange}
      worktreeId={worktreeId}
      repositoryId={repositoryId}
      preferRepositoryHint={preferRepositoryHint}
      hideLabel
    />
  );
}
