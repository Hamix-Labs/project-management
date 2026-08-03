import { PromptEditorSaveStatus, type PromptEditorSaveStatusKind } from "./PromptEditorSaveStatus";

export type PromptEditorTopbarProps = {
  backLabel?: string;
  title: string;
  saveStatus: PromptEditorSaveStatusKind;
  saveErrorDetail?: string;
  leavePending?: boolean;
  onBack: () => void;
  onRetrySave?: () => void;
};

export function PromptEditorTopbar({
  backLabel = "Back to task",
  title,
  saveStatus,
  saveErrorDetail,
  leavePending = false,
  onBack,
  onRetrySave,
}: PromptEditorTopbarProps) {
  return (
    <header className="prompt-editor-topbar">
      <div className="prompt-editor-topbar__left">
        <button
          type="button"
          className="prompt-editor-topbar__back"
          disabled={leavePending}
          onClick={onBack}
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden="true"
          >
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
          {backLabel}
        </button>
        <span className="prompt-editor-topbar__sep" aria-hidden="true">
          /
        </span>
        <div className="prompt-editor-topbar__crumb">
          <b>{title}</b>
        </div>
      </div>
      <div className="prompt-editor-topbar__right">
        <PromptEditorSaveStatus
          kind={saveStatus}
          errorDetail={saveErrorDetail}
          onRetry={onRetrySave}
        />
        <button
          type="button"
          className="prompt-editor-topbar__history"
          aria-disabled="true"
          title="Version history (coming soon)"
          tabIndex={-1}
        >
          <svg
            width="17"
            height="17"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
            aria-hidden="true"
          >
            <path d="M3 12a9 9 0 1 0 2.6-6.3L3 8" />
            <path d="M3 3v5h5" />
            <path d="M12 7v5l3 3" />
          </svg>
          <span className="visually-hidden">Version history (coming soon)</span>
        </button>
      </div>
    </header>
  );
}
