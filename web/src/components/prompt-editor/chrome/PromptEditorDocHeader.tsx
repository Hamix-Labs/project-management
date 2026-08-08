import { useEffect, useRef, useState } from "react";

export type PromptEditorDocHeaderProps = {
  title: string;
  editedLabel: string;
  wordCountLabel: string;
  repoLabel: string;
  disabled?: boolean;
  onTitleCommit?: (next: string) => void | Promise<void>;
};

export function PromptEditorDocHeader({
  title,
  editedLabel,
  wordCountLabel,
  repoLabel,
  disabled = false,
  onTitleCommit,
}: PromptEditorDocHeaderProps) {
  const [draft, setDraft] = useState(title);
  const committedRef = useRef(title);
  const skipBlurCommitRef = useRef(false);

  useEffect(() => {
    setDraft(title);
    committedRef.current = title;
  }, [title]);

  const editable = Boolean(onTitleCommit) && !disabled;

  const commit = () => {
    if (!onTitleCommit || disabled) return;
    const next = draft.trim();
    if (!next) {
      setDraft(committedRef.current);
      return;
    }
    if (next === committedRef.current) {
      setDraft(committedRef.current);
      return;
    }
    void onTitleCommit(next);
  };

  return (
    <div className="prompt-editor-doc-header">
      {editable ? (
        <input
          type="text"
          className="prompt-editor-doc-header__title"
          aria-label="Document title"
          value={draft}
          disabled={disabled}
          onChange={(ev) => setDraft(ev.target.value)}
          onBlur={() => {
            if (skipBlurCommitRef.current) {
              skipBlurCommitRef.current = false;
              return;
            }
            commit();
          }}
          onKeyDown={(ev) => {
            if (ev.key === "Enter") {
              ev.preventDefault();
              skipBlurCommitRef.current = true;
              (ev.target as HTMLInputElement).blur();
              commit();
              return;
            }
            if (ev.key === "Escape") {
              ev.preventDefault();
              ev.stopPropagation();
              skipBlurCommitRef.current = true;
              setDraft(committedRef.current);
              (ev.target as HTMLInputElement).blur();
            }
          }}
        />
      ) : (
        <h1 className="prompt-editor-doc-header__title">{title}</h1>
      )}
      <div className="prompt-editor-doc-header__meta">
        <span className="prompt-editor-doc-header__meta-item">
          <svg
            width="13"
            height="13"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden="true"
          >
            <rect x="3" y="4" width="18" height="18" rx="2" />
            <path d="M16 2v4M8 2v4M3 10h18" />
          </svg>
          {editedLabel}
        </span>
        <span className="prompt-editor-doc-header__meta-item">
          <svg
            width="13"
            height="13"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden="true"
          >
            <path d="M9 17H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2h-4" />
            <path d="M9 21h6M12 17v4" />
          </svg>
          {wordCountLabel}
        </span>
        <span className="prompt-editor-doc-header__meta-item">
          <svg
            width="13"
            height="13"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden="true"
          >
            <path d="M4 20V10M12 20V4M20 20v-7" />
          </svg>
          {repoLabel}
        </span>
      </div>
    </div>
  );
}
