import { FieldLabel } from "@/shared/FieldLabel";
import { RichPromptEditor } from "@/components/rich-prompt";

type Props = {
  idsPrefix: string;
  editorKey: string;
  prompt: string;
  disabled: boolean;
  onPromptChange: (v: string) => void;
  worktreeId?: string;
  /** When true, the section header owns the label — omit the field label. */
  hideLabel?: boolean;
};

export function TaskComposePromptField({
  idsPrefix,
  editorKey,
  prompt,
  disabled,
  onPromptChange,
  worktreeId,
  hideLabel = false,
}: Props) {
  const promptId = `${idsPrefix}-prompt`;

  return (
    <div className="field grow stack-tight prompt-field-full task-create-prompt">
      {hideLabel ? null : (
        <FieldLabel
          id={`${promptId}-label`}
          htmlFor={promptId}
          requirement="optional"
        >
          Initial prompt
        </FieldLabel>
      )}
      <div className="task-create-editor-shell">
        <RichPromptEditor
          key={editorKey}
          id={promptId}
          value={prompt}
          onChange={onPromptChange}
          disabled={disabled}
          placeholder="Describe the task in detail. Type @ to mention a repo file…"
          worktreeId={worktreeId}
        />
      </div>
    </div>
  );
}
