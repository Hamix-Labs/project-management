export type PromptEditorDocHeaderProps = {
  title: string;
  badgeLabel?: string;
  editedLabel: string;
  wordCountLabel: string;
  repoLabel: string;
};

export function PromptEditorDocHeader({
  title,
  badgeLabel = "Implementation brief",
  editedLabel,
  wordCountLabel,
  repoLabel,
}: PromptEditorDocHeaderProps) {
  return (
    <div className="prompt-editor-doc-header">
      <p className="prompt-editor-doc-header__eyebrow">
        <span className="prompt-editor-doc-header__badge">{badgeLabel}</span>
      </p>
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
            <path d="M4 20V10M12 20V4M20 20v-7" />
          </svg>
          {repoLabel}
        </span>
      </div>
    </div>
  );
}
