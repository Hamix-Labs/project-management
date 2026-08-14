import { useCallback, useEffect, useRef, useState } from "react";
import {
  DRAFT_ASSIST_WATCHDOG_OFFER_CANCEL_MS,
  useDraftAssistWatchdog,
} from "./useDraftAssistWatchdog";
import {
  draftAssistStatusCopy,
  isDraftAssistRunActive,
} from "./draftAssistStatus";
import { useDraftAssistContext } from "./DraftAssistContext";
import { DraftAssistMessage } from "./DraftAssistMessage";

/**
 * Live thread rendered in the compose page's reserved assist column.
 * Consumes {@link DraftAssistContext} — the compose form wraps
 * children in `DraftAssistProvider`, then mounts this panel when the
 * operator triggers Space-for-AI.
 *
 * A11y (design §Accessibility):
 * - Primary status region is `aria-live="polite"`.
 * - Errors are announced `assertive` via their own region.
 * - Stop button stays keyboard-reachable (regular <button>).
 */
export function DraftAssistThread() {
  const ctx = useDraftAssistContext();
  const [input, setInput] = useState("");
  const listRef = useRef<HTMLUListElement | null>(null);

  const watchdog = useDraftAssistWatchdog(
    ctx.runActive,
    ctx.runStartedAt,
  );

  const statusText = draftAssistStatusCopy(ctx.status);
  const running = isDraftAssistRunActive(ctx.status);
  const showCancel =
    running || watchdog.elapsedMs >= DRAFT_ASSIST_WATCHDOG_OFFER_CANCEL_MS;

  useEffect(() => {
    // Auto-scroll on new messages / streaming tokens — modest surface,
    // no fancy virtualization needed for a per-compose thread.
    const el = listRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [ctx.messages]);

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      const trimmed = input.trim();
      if (trimmed === "") return;
      if (ctx.runActive) return;
      ctx.send(trimmed);
      setInput("");
    },
    [ctx, input],
  );

  const handleStop = useCallback(() => {
    ctx.stop();
  }, [ctx]);

  const handleRetry = useCallback(() => {
    // Retry means: re-open with the last user message. If none, just
    // re-arm the session so the operator can send fresh.
    const lastUser = [...ctx.messages].reverse().find((m) => m.kind === "user");
    if (lastUser && lastUser.kind === "user") {
      ctx.send(lastUser.text);
    }
  }, [ctx]);

  return (
    <section
      className="draft-assist-thread"
      aria-label="Draft assistant"
      data-testid="draft-assist-thread"
    >
      <header className="draft-assist-thread__header">
        <h2 className="draft-assist-thread__title">Draft assistant</h2>
        <p className="draft-assist-thread__hint">
          Ask the assistant to tighten this brief. Prompt edits apply
          in place — hit Create when you’re ready.
        </p>
      </header>

      <ul
        ref={listRef}
        className="draft-assist-thread__messages"
        aria-label="Assistant messages"
        aria-live="off"
      >
        {ctx.messages.map((m) => (
          <DraftAssistMessage key={m.id} message={m} />
        ))}
      </ul>

      <div
        className="draft-assist-thread__status"
        aria-live="polite"
        role="status"
        data-status={ctx.status.status}
      >
        {statusText ? (
          <span className="draft-assist-thread__status-text">{statusText}</span>
        ) : null}
        {watchdog.message ? (
          <span className="draft-assist-thread__watchdog">
            {watchdog.message}
          </span>
        ) : null}
      </div>

      {ctx.status.errorMessage ? (
        <div
          className="draft-assist-thread__error"
          role="alert"
          aria-live="assertive"
        >
          {ctx.status.errorMessage}
        </div>
      ) : null}

      <form
        className="draft-assist-thread__composer"
        onSubmit={handleSubmit}
      >
        <label
          htmlFor="draft-assist-thread-input"
          className="visually-hidden"
        >
          Message the assistant
        </label>
        <textarea
          id="draft-assist-thread-input"
          className="draft-assist-thread__input"
          value={input}
          rows={2}
          placeholder="Ask a follow-up…"
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              handleSubmit(e as unknown as React.FormEvent);
            }
          }}
        />
        <div className="draft-assist-thread__actions">
          {running ? (
            <button
              type="button"
              className="draft-assist-thread__stop"
              onClick={handleStop}
              data-testid="draft-assist-stop"
            >
              Stop
            </button>
          ) : (
            <button
              type="submit"
              className="draft-assist-thread__send"
              disabled={input.trim() === ""}
              data-testid="draft-assist-send"
            >
              Send
            </button>
          )}
          {!running && showCancel ? (
            <button
              type="button"
              className="draft-assist-thread__stop"
              onClick={handleStop}
              data-testid="draft-assist-stop-late"
            >
              Cancel
            </button>
          ) : null}
          {watchdog.offerRetry ? (
            <button
              type="button"
              className="draft-assist-thread__retry"
              onClick={handleRetry}
              data-testid="draft-assist-retry"
            >
              Retry
            </button>
          ) : null}
        </div>
      </form>
    </section>
  );
}
