import type { RichPromptEditorProjectContextProps } from "../../rich-prompt";
import { TaskComposePromptField } from "../../task-compose/fields/TaskComposePromptField";

type Props = {
  idsPrefix: string;
  editorKey: string;
  prompt: string;
  disabled: boolean;
  onPromptChange: (v: string) => void;
  projectContext?: RichPromptEditorProjectContextProps;
  worktreeId?: string;
};

export function TaskCreateModalPromptFields({
  idsPrefix,
  editorKey,
  prompt,
  disabled,
  onPromptChange,
  projectContext,
  worktreeId,
}: Props) {
  return (
    <TaskComposePromptField
      idsPrefix={idsPrefix}
      editorKey={editorKey}
      prompt={prompt}
      disabled={disabled}
      onPromptChange={onPromptChange}
      projectContext={projectContext}
      worktreeId={worktreeId}
      hideLabel
    />
  );
}
