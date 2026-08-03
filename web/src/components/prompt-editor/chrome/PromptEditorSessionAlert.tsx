export type PromptEditorSessionAlertProps = {
  title: string;
  detail: string;
  variant?: "error" | "warning";
  onRetry?: () => void;
  onBack?: () => void;
  onDismiss?: () => void;
  retryLabel?: string;
  backLabel?: string;
};

/**
 * Actionable session alert: title, detail, and optional Retry / Back / Dismiss.
 */
export function PromptEditorSessionAlert({
  title,
  detail,
  variant = "error",
  onRetry,
  onBack,
  onDismiss,
  retryLabel = "Retry",
  backLabel = "Back",
}: PromptEditorSessionAlertProps) {
  const rootClass =
    variant === "warning"
      ? "prompt-editor-session-alert prompt-editor-session-alert--warning"
      : "prompt-editor-session-alert prompt-editor-session-alert--error err";

  return (
    <div className={rootClass} role="alert">
      <p className="prompt-editor-session-alert__title">{title}</p>
      <p className="prompt-editor-session-alert__detail">{detail}</p>
      {onRetry || onBack || onDismiss ? (
        <div className="prompt-editor-session-alert__actions">
          {onRetry ? (
            <button type="button" className="primary" onClick={onRetry}>
              {retryLabel}
            </button>
          ) : null}
          {onBack ? (
            <button type="button" className="ghost" onClick={onBack}>
              {backLabel}
            </button>
          ) : null}
          {onDismiss ? (
            <button type="button" className="ghost" onClick={onDismiss}>
              Dismiss
            </button>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
