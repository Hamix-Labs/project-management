import { forwardRef } from "react";

type Props = {
  open: boolean;
  disabled?: boolean;
  onToggle: () => void;
};

/**
 * Header-mounted trigger that opens the sample-task popover. Lives in its
 * own component so the create modal can keep a clean ref to the underlying
 * button (the popover anchors to it via getBoundingClientRect).
 */
export const TestScenariosTrigger = forwardRef<HTMLButtonElement, Props>(
  function TestScenariosTrigger({ open, disabled, onToggle }, ref) {
    return (
      <button
        ref={ref}
        type="button"
        className="test-scenarios-trigger"
        data-testid="test-scenarios-trigger"
        data-active={open ? "true" : "false"}
        aria-haspopup="dialog"
        aria-expanded={open}
        disabled={disabled}
        onClick={onToggle}
      >
        <FlaskGlyph />
        <span className="test-scenarios-trigger__label">Test scenarios</span>
      </button>
    );
  },
);

function FlaskGlyph() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M14 2v6a2 2 0 0 0 .245.96l5.51 10.08A2 2 0 0 1 18 22H6a2 2 0 0 1-1.755-2.96l5.51-10.08A2 2 0 0 0 10 8V2" />
      <path d="M6.453 15h11.094" />
      <path d="M8.5 2h7" />
    </svg>
  );
}
