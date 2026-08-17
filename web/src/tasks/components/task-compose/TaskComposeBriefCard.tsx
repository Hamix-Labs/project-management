import { useCallback } from "react";
import { RichPromptEditor } from "@/components/rich-prompt";
import { useOptionalDraftAssistContext } from "@/tasks/components/draft-assist";
import { useComposeBriefVerticalResize } from "./useComposeBriefVerticalResize";
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

/** Empty-block cursor copy; `/` opens the existing rich-prompt slash menu. */
export const COMPOSE_BRIEF_PLACEHOLDER = "Press `/` for commands";

/**
 * Brief hero card: large title input + rich prompt editor.
 * Structure is inserted via `/` slash commands, not a markdown block toolbar.
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
  const {
    rootRef,
    onGripPointerDown,
    onGripPointerMove,
    onGripPointerUp,
    onGripPointerCancel,
  } = useComposeBriefVerticalResize();
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
    <section
      ref={rootRef}
      className="compose-card compose-brief"
      aria-labelledby={titleId}
    >
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
          placeholder={COMPOSE_BRIEF_PLACEHOLDER}
          worktreeId={worktreeId}
          repositoryId={repositoryId}
          preferRepositoryHint={preferRepositoryHint}
          menuVariant="none"
          menuRight={
            <span className="compose-brief__word-count" aria-live="polite">
              {wordCount} words
            </span>
          }
          onEditorReady={onEditorReady}
          onAiTrigger={draftAssist ? onAiTrigger : undefined}
        />
      </div>
      {/* Pointer-only, matching native textarea resize; editor keeps keyboard focus. */}
      <div
        className="compose-brief__resize-grip"
        aria-hidden="true"
        onPointerDown={onGripPointerDown}
        onPointerMove={onGripPointerMove}
        onPointerUp={onGripPointerUp}
        onPointerCancel={onGripPointerCancel}
        onLostPointerCapture={onGripPointerCancel}
      >
        <svg viewBox="0 0 12 12" width="12" height="12" focusable="false">
          <path d="M4 12 L12 4" />
          <path d="M8.5 12 L12 8.5" />
        </svg>
      </div>
    </section>
  );
}
