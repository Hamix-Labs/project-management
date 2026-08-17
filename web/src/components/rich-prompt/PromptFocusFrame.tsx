import { useCallback, useRef, type ReactNode, type RefObject } from "react";
import { usePromptFocus } from "./usePromptFocus";

export type PromptFocusFrameProps = {
  expanded: boolean;
  onExpandedChange: (next: boolean) => void;
  label: string;
  wordCount: number;
  disabled?: boolean;
  title: ReactNode;
  children: ReactNode;
  restoreFocusRef: RefObject<HTMLElement | null>;
};

function CollapseIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" aria-hidden="true" focusable="false">
      <path
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        d="M8 3v5H3M16 3v5h5M8 21v-5H3M16 21v-5h5"
      />
    </svg>
  );
}

/** Same editor instance; expanded state only restyles the wrapping frame. */
export function PromptFocusFrame({
  expanded,
  onExpandedChange,
  label,
  wordCount,
  disabled,
  title,
  children,
  restoreFocusRef,
}: PromptFocusFrameProps) {
  const frameRef = useRef<HTMLDivElement>(null);
  const onClose = useCallback(() => onExpandedChange(false), [onExpandedChange]);
  usePromptFocus({
    expanded,
    onClose,
    frameRef,
    restoreFocusRef,
  });

  return (
    <div
      ref={frameRef}
      className={
        expanded ? "prompt-focus-frame prompt-focus-frame--expanded" : "prompt-focus-frame"
      }
      role={expanded ? "dialog" : undefined}
      aria-modal={expanded ? true : undefined}
      aria-label={expanded ? label : undefined}
    >
      {expanded ? (
        <header className="prompt-focus-frame__header">
          <span className="prompt-focus-frame__label">{label}</span>
          <span className="prompt-focus-frame__count">{wordCount} words</span>
          <button
            type="button"
            className="prompt-focus-frame__done"
            onClick={() => onExpandedChange(false)}
            disabled={disabled}
          >
            <CollapseIcon />
            Done
          </button>
        </header>
      ) : null}
      <div className="prompt-focus-frame__body">
        {title}
        {children}
      </div>
    </div>
  );
}
