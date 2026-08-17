import { useCallback, useRef, useState } from "react";
import { PromptFocusFrame, RichPromptEditor } from "@/components/rich-prompt";
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

/** Empty-block cursor copy; `/` opens the Notion slash menu (Hamix Ask AI + files included). */
export const COMPOSE_BRIEF_PLACEHOLDER = "Press `/` for commands";

function ExpandIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" aria-hidden="true" focusable="false">
      <path
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        d="M15 3h6v6M9 21H3v-6M21 3l-7 7M3 21l7-7"
      />
    </svg>
  );
}

/** Brief hero card: title + word count + Notion editor + focused writing overlay. */
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
  const expandRef = useRef<HTMLButtonElement>(null);
  const [expanded, setExpanded] = useState(false);
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
      data-focus-expanded={expanded ? "true" : undefined}
    >
      <PromptFocusFrame
        expanded={expanded}
        onExpandedChange={setExpanded}
        label="Editing brief"
        wordCount={wordCount}
        disabled={disabled}
        restoreFocusRef={expandRef}
        title={
          <div className="compose-brief__title-block">
            <div className="compose-brief__title-row">
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
              <button
                ref={expandRef}
                type="button"
                className="compose-brief__expand"
                aria-label="Expand"
                onClick={() => setExpanded(true)}
                disabled={disabled}
              >
                <ExpandIcon />
              </button>
              <span className="compose-brief__word-count" aria-live="polite">
                {wordCount} words
              </span>
            </div>
          </div>
        }
      >
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
            onEditorReady={onEditorReady}
            onAiTrigger={draftAssist ? onAiTrigger : undefined}
          />
        </div>
      </PromptFocusFrame>
      <div
        className="compose-brief__resize-grip"
        aria-hidden="true"
        onPointerDown={onGripPointerDown}
        onPointerMove={onGripPointerMove}
        onPointerUp={onGripPointerUp}
        onPointerCancel={onGripPointerCancel}
        onLostPointerCapture={onGripPointerCancel}
      >
        <svg viewBox="0 0 16 16" width="16" height="16" focusable="false">
          <path d="M6 14 L14 6" />
          <path d="M10 15 L15 10" />
        </svg>
      </div>
    </section>
  );
}
