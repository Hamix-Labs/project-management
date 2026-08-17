import { useCallback, type ReactNode } from "react";
import { RichPromptEditor } from "@/components/rich-prompt";
import { useOptionalDraftAssistContext } from "@/tasks/components/draft-assist";
import { useRichPromptWordCount } from "./useRichPromptWordCount";

type Props = {
  idsPrefix: string;
  editorKey: string;
  title: string;
  prompt: string;
  disabled: boolean;
  worktreeId?: string;
  repositoryId?: string;
  preferRepositoryHint?: boolean;
  onTitleChange: (v: string) => void;
  onPromptChange: (v: string) => void;
};

/**
 * Brief hero card: large title input + rich prompt editor with icon toolbar.
 * Bound to the same essentials/prompt handlers as the former essentials+prompt sections.
 */
export function TaskComposeBriefCard({
  idsPrefix,
  editorKey,
  title,
  prompt,
  disabled,
  worktreeId,
  repositoryId,
  preferRepositoryHint = false,
  onTitleChange,
  onPromptChange,
}: Props) {
  const titleId = `${idsPrefix}-title`;
  const promptId = `${idsPrefix}-prompt`;
  const { wordCount, onEditorReady } = useRichPromptWordCount();
  const draftAssist = useOptionalDraftAssistContext();
  const onAiTrigger = useCallback(
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

  return (
    <section className="compose-card compose-brief" aria-labelledby={titleId}>
      <div className="compose-brief__title-block">
        <input
          id={titleId}
          className="compose-brief__title-input"
          value={title}
          onChange={(e) => onTitleChange(e.target.value)}
          placeholder="Name this task"
          required
          aria-required="true"
          aria-label="Title"
          disabled={disabled}
        />
      </div>
      <div className="compose-brief__editor task-create-editor-shell">
        <RichPromptEditor
          key={editorKey}
          id={promptId}
          value={prompt}
          onChange={onPromptChange}
          disabled={disabled}
          placeholder="Describe the full brief the agent starts from. Supports Markdown."
          worktreeId={worktreeId}
          repositoryId={repositoryId}
          preferRepositoryHint={preferRepositoryHint}
          menuVariant="icons"
          menuRight={
            <span className="compose-brief__word-count" aria-live="polite">
              {wordCount} words
            </span>
          }
          onEditorReady={onEditorReady}
          onAiTrigger={draftAssist ? onAiTrigger : undefined}
        />
      </div>
    </section>
  );
}

/** Shared section title chrome for rail cards. */
export function ComposeRailSectionTitle({
  icon,
  children,
}: {
  icon: ReactNode;
  children: ReactNode;
}) {
  return (
    <div className="compose-card__section-title">
      <span className="compose-card__section-icon">{icon}</span>
      <h2 className="compose-card__section-label">{children}</h2>
    </div>
  );
}
