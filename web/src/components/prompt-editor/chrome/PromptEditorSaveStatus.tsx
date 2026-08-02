export type PromptEditorSaveStatusKind =
  | "saved"
  | "saving"
  | "unsaved"
  | "error";

export type PromptEditorSaveStatusProps = {
  kind: PromptEditorSaveStatusKind;
  onRetry?: () => void;
};

export function PromptEditorSaveStatus({
  kind,
  onRetry,
}: PromptEditorSaveStatusProps) {
  if (kind === "saving") {
    return (
      <span className="prompt-editor-save-status" aria-live="polite">
        <span
          className="prompt-editor-save-status__dot prompt-editor-save-status__dot--muted"
          aria-hidden="true"
        />
        Saving…
      </span>
    );
  }
  if (kind === "error") {
    return (
      <span
        className="prompt-editor-save-status prompt-editor-save-status--error"
        aria-live="assertive"
      >
        <span
          className="prompt-editor-save-status__dot prompt-editor-save-status__dot--error"
          aria-hidden="true"
        />
        Couldn&apos;t save
        {onRetry ? (
          <button
            type="button"
            className="prompt-editor-save-status__retry"
            onClick={onRetry}
          >
            Retry
          </button>
        ) : null}
      </span>
    );
  }
  if (kind === "unsaved") {
    return (
      <span className="prompt-editor-save-status" aria-live="polite">
        <span
          className="prompt-editor-save-status__dot prompt-editor-save-status__dot--muted"
          aria-hidden="true"
        />
        Unsaved changes
      </span>
    );
  }
  return (
    <span className="prompt-editor-save-status" aria-live="polite">
      <span className="prompt-editor-save-status__dot" aria-hidden="true" />
      Saved to draft
    </span>
  );
}
