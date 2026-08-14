import { FieldLabel } from "@/shared/FieldLabel";
import { RichPromptEditor } from "@/components/rich-prompt";

type Props = {
  idsPrefix: string;
  editorKey: string;
  prompt: string;
  disabled: boolean;
  onPromptChange: (v: string) => void;
  worktreeId?: string;
  repositoryId?: string;
  /** Create flow: hint asks for a repository when none is selected. */
  preferRepositoryHint?: boolean;
  /** When true, the section header owns the label — omit the field label. */
  hideLabel?: boolean;
  /**
   * Optional callback fired when the operator invokes Space-for-AI or picks
   * `/ai` from the slash menu. Default no-op in Plan 1; Plan 3 wires this
   * to `hamix-draft-agent`.
   */
  onAiTrigger?: (msg: string) => void;
};

const NOOP_AI_TRIGGER = (_msg: string) => {
  void _msg;
};

export function TaskComposePromptField({
  idsPrefix,
  editorKey,
  prompt,
  disabled,
  onPromptChange,
  worktreeId,
  repositoryId,
  preferRepositoryHint = false,
  hideLabel = false,
  onAiTrigger = NOOP_AI_TRIGGER,
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
          worktreeId={worktreeId}
          repositoryId={repositoryId}
          preferRepositoryHint={preferRepositoryHint}
          onAiTrigger={onAiTrigger}
        />
      </div>
    </div>
  );
}
