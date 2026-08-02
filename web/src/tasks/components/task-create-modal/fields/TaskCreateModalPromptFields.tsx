import { TaskComposePromptField } from "../../task-compose/fields/TaskComposePromptField";

type Props = {
  idsPrefix: string;
  prompt: string;
  disabled: boolean;
  onOpenPromptEditor: () => void;
};

export function TaskCreateModalPromptFields({
  idsPrefix,
  prompt,
  disabled,
  onOpenPromptEditor,
}: Props) {
  return (
    <TaskComposePromptField
      idsPrefix={idsPrefix}
      prompt={prompt}
      disabled={disabled}
      onOpenPromptEditor={onOpenPromptEditor}
    />
  );
}
