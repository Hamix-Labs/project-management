import type { DraftAssistThreadMessage } from "./DraftAssistContext";

type Props = {
  message: DraftAssistThreadMessage;
};

/**
 * One row in the assist thread. Renders per-kind so styling and
 * semantics stay specific (user bubble, assistant streaming, tool
 * event, patch note, error banner). Content is plain text — SSE
 * tokens are model output, never HTML.
 */
export function DraftAssistMessage({ message }: Props) {
  switch (message.kind) {
    case "user":
      return (
        <li
          className="draft-assist-message draft-assist-message--user"
          data-testid="draft-assist-user-message"
        >
          <div className="draft-assist-message__role">You</div>
          <div className="draft-assist-message__body">{message.text}</div>
        </li>
      );

    case "assistant":
      return (
        <li
          className="draft-assist-message draft-assist-message--assistant"
          data-testid="draft-assist-assistant-message"
          data-done={message.done ? "true" : "false"}
        >
          <div className="draft-assist-message__role">Assistant</div>
          <div className="draft-assist-message__body">
            {message.text}
            {message.done ? null : (
              <span
                className="draft-assist-message__caret"
                aria-hidden="true"
              />
            )}
          </div>
        </li>
      );

    case "tool": {
      const label =
        message.phase === "start"
          ? `Reading ${message.name}…`
          : message.ok === false
            ? `${message.name} failed${message.error ? `: ${message.error}` : ""}`
            : `${message.name} done`;
      return (
        <li
          className="draft-assist-message draft-assist-message--tool"
          data-testid="draft-assist-tool-row"
          data-phase={message.phase}
        >
          <span className="draft-assist-message__glyph" aria-hidden="true">
            ⚙
          </span>
          <span className="draft-assist-message__tool-label">{label}</span>
        </li>
      );
    }

    case "patch": {
      const label = message.applied
        ? message.summary
          ? `Prompt updated — ${message.summary}`
          : "Prompt updated"
        : "Prompt update skipped (invalid patch)";
      return (
        <li
          className="draft-assist-message draft-assist-message--patch"
          data-testid="draft-assist-patch-row"
          data-op={message.op}
          data-applied={message.applied ? "true" : "false"}
        >
          <span className="draft-assist-message__glyph" aria-hidden="true">
            ✎
          </span>
          <span className="draft-assist-message__patch-label">{label}</span>
        </li>
      );
    }

    case "error":
      return (
        <li
          className="draft-assist-message draft-assist-message--error"
          data-testid="draft-assist-error-row"
          role="alert"
        >
          <span className="draft-assist-message__glyph" aria-hidden="true">
            ⚠
          </span>
          <span className="draft-assist-message__error-text">
            {message.message}
          </span>
        </li>
      );
  }
}
