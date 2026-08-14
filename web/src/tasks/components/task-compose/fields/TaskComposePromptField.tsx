import { useCallback } from "react";
import { FieldLabel } from "@/shared/FieldLabel";
import { RichPromptEditor } from "@/components/rich-prompt";
import { useOptionalDraftAssistContext } from "@/tasks/components/draft-assist";

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
   * `/ai` from the slash menu. When omitted, the field falls back to the
   * nearest `DraftAssistProvider`: the first trigger calls `open(msg)` so
   * Plan 3's assist thread opens lazily; subsequent triggers reuse the
   * session via `send(msg)`.
   */
  onAiTrigger?: (msg: string) => void;
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
  onAiTrigger,
}: Props) {
  const draftAssist = useOptionalDraftAssistContext();
  const contextTrigger = useCallback(
    (msg: string) => {
      if (!draftAssist) return;
      if (draftAssist.active) {
        draftAssist.send(msg);
      } else {
        draftAssist.open(msg);
      }
    },
    [draftAssist],
  );
  const trigger = onAiTrigger ?? (draftAssist ? contextTrigger : undefined);
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
          onAiTrigger={trigger}
        />
      </div>
    </div>
  );
}
