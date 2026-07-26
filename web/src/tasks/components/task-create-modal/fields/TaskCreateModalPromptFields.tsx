import { TaskComposePromptField } from "../../task-compose/fields/TaskComposePromptField";

type Props = {
  idsPrefix: string;
  editorKey: string;
  prompt: string;
  disabled: boolean;
  onPromptChange: (v: string) => void;
  worktreeId?: string;
};

export function TaskCreateModalPromptFields({
  idsPrefix,
  editorKey,
  prompt,
  disabled,
  onPromptChange,
  worktreeId,
}: Props) {
  return (
    <TaskComposePromptField
      idsPrefix={idsPrefix}
      editorKey={editorKey}
      prompt={prompt}
      disabled={disabled}
      onPromptChange={onPromptChange}
      worktreeId={worktreeId}
      hideLabel
    />
  );
}
