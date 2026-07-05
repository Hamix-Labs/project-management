import { FieldLabel } from "@/shared/FieldLabel";
import {
  RichPromptEditor,
  type RichPromptEditorProjectContextProps,
} from "../../rich-prompt";

type Props = {
  idsPrefix: string;
  editorKey: string;
  prompt: string;
  disabled: boolean;
  onPromptChange: (v: string) => void;
  projectContext?: RichPromptEditorProjectContextProps;
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
  projectContext,
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
          placeholder={
            projectContext
              ? "Describe the task in detail. Type # to reference project context…"
              : "Describe the task in detail. Type @ to mention a repo file…"
          }
          projectContext={projectContext}
          worktreeId={worktreeId}
        />
      </div>
    </div>
  );
}
