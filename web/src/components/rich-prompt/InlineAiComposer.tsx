import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import tippy, { type Instance as TippyInstance } from "tippy.js";

export type InlineAiComposerProps = {
  open: boolean;
  /** Called to compute the reference rect for the tippy anchor. */
  getAnchorRect: () => DOMRect | null;
  /** Initial value seeded into the input (e.g. from `/ai <query>`). */
  initialValue?: string;
  onClose: () => void;
  onSubmit: (msg: string) => void;
};

const PLACEHOLDER = "Ask the assistant to help with this brief…";

/** Media-query check pulled out for tests to stub `matchMedia`. */
function prefersReducedMotion(): boolean {
  if (typeof window === "undefined") return false;
  if (typeof window.matchMedia !== "function") return false;
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

/**
 * Inline composer shell for Space-for-AI. Anchored via tippy to the empty
 * block that triggered it. Plan 3 wires Enter to a real send; here Enter
 * fires `onSubmit(msg)` and Escape closes the popover.
 */
export function InlineAiComposer({
  open,
  getAnchorRect,
  initialValue = "",
  onClose,
  onSubmit,
}: InlineAiComposerProps) {
  const [value, setValue] = useState(initialValue);
  const [host, setHost] = useState<HTMLDivElement | null>(null);
  const popupRef = useRef<TippyInstance | null>(null);
  const inputRef = useRef<HTMLTextAreaElement | null>(null);

  useEffect(() => {
    if (!open) return;
    setValue(initialValue);
  }, [open, initialValue]);

  useEffect(() => {
    if (!open) {
      setHost(null);
      return;
    }
    const el = document.createElement("div");
    el.className = "inline-ai-composer__host";
    document.body.appendChild(el);
    setHost(el);

    const reduced = prefersReducedMotion();
    const inst = tippy(document.body, {
      getReferenceClientRect: () => getAnchorRect() ?? new DOMRect(0, 0, 0, 0),
      appendTo: () => document.body,
      content: el,
      showOnCreate: true,
      interactive: true,
      trigger: "manual",
      placement: "bottom-start",
      theme: "inline-ai-composer",
      arrow: false,
      offset: [0, 6],
      duration: reduced ? 0 : [140, 100],
      animation: reduced ? undefined : "shift-away-subtle",
    });
    popupRef.current = Array.isArray(inst) ? inst[0]! : inst;

    return () => {
      popupRef.current?.destroy();
      popupRef.current = null;
      el.remove();
      setHost(null);
    };
  }, [open, getAnchorRect]);

  useEffect(() => {
    if (!open || !host) return;
    const raf = window.requestAnimationFrame(() => {
      inputRef.current?.focus();
    });
    return () => window.cancelAnimationFrame(raf);
  }, [open, host]);

  if (!open || !host) return null;

  const surface = (
    <div
      className="inline-ai-composer"
      role="dialog"
      aria-label="AI assistant composer"
    >
      <textarea
        ref={inputRef}
        className="inline-ai-composer__input"
        value={value}
        placeholder={PLACEHOLDER}
        rows={2}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Escape") {
            e.preventDefault();
            e.stopPropagation();
            onClose();
            return;
          }
          if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            onSubmit(value);
          }
        }}
        aria-label="Message the assistant"
      />
      <div className="inline-ai-composer__footer" aria-hidden="true">
        <span className="inline-ai-composer__hint">
          <kbd>Enter</kbd> send · <kbd>Shift+Enter</kbd> newline ·{" "}
          <kbd>Esc</kbd> close
        </span>
      </div>
    </div>
  );

  return createPortal(surface, host);
}
