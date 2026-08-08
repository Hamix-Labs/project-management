export type PromptEditorDocHeaderProps = {
  title: string;
  editedLabel: string;
  wordCountLabel: string;
  repoLabel: string;
};

export function PromptEditorDocHeader({
  title,
  editedLabel,
  wordCountLabel,
  repoLabel,
}: PromptEditorDocHeaderProps) {
  return (
    <div className="prompt-editor-doc-header">
      <h1 className="prompt-editor-doc-header__title">{title}</h1>
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
            <path d="M4 7a2 2 0 0 1 2-2h3.5l2 2H18a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V7Z" />
            <circle cx="10.5" cy="13.5" r="1.75" />
            <path d="M12.25 13.5H16" />
          </svg>
          <span className="visually-hidden">Repository: </span>
          {repoLabel}
        </span>
      </div>
    </div>
  );
}
