import { useEffect, useId, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { TEST_SCENARIOS, type TestScenario } from "@/tasks/test-scenarios";

type Props = {
  /**
   * Trigger button the popover anchors to. Re-positioned on viewport
   * scroll / resize so the popover tracks the trigger when the surrounding
   * modal scrolls. Pass `null` while the trigger is unmounted.
   */
  anchor: HTMLElement | null;
  onPick: (scenario: TestScenario) => void;
  onClose: () => void;
};

const POPOVER_VERTICAL_GAP = 8;
const POPOVER_VIEWPORT_MARGIN = 12;
const POPOVER_DEFAULT_WIDTH = 320;
/**
 * Same z tier as the schedule quick-pick popover (see
 * `QuickScheduleOffsetPopover`) so both float above the create modal
 * shell. Only one trigger is ever open per modal so they cannot overlap.
 */
const POPOVER_Z_INDEX = 13000;

/** Survives popover remount so the last-picked check shows on reopen. Not persisted. */
let lastPickedScenarioId: string | null = null;

/**
 * Anchored popover listing the three sample tasks. Picking a row calls
 * `onPick` (the parent applies it to the form) and closes the popover.
 *
 * Renders into `document.body` so the modal's overflow does not clip it.
 * Escape closes and returns focus to the trigger; click-outside dismisses.
 */
export function TestScenariosPopover({ anchor, onPick, onClose }: Props) {
  const titleId = useId();
  const popoverRef = useRef<HTMLDivElement>(null);
  const [lastPickedId, setLastPickedId] = useState<string | null>(
    lastPickedScenarioId,
  );
  const [pos, setPos] = useState<{
    top: number;
    left: number;
    width: number;
    placeAbove: boolean;
  } | null>(null);

  useLayoutEffect(() => {
    if (!anchor) {
      setPos(null);
      return;
    }
    const compute = () => {
      const rect = anchor.getBoundingClientRect();
      const popHeight = popoverRef.current?.offsetHeight ?? 0;
      const viewportH = window.innerHeight;
      const viewportW = window.innerWidth;
      const spaceBelow = viewportH - rect.bottom - POPOVER_VERTICAL_GAP;
      const placeAbove =
        popHeight > 0 &&
        spaceBelow < popHeight &&
        rect.top - POPOVER_VERTICAL_GAP > popHeight;
      const width = Math.min(
        POPOVER_DEFAULT_WIDTH,
        viewportW - 2 * POPOVER_VIEWPORT_MARGIN,
      );
      // Right-align the popover with the trigger (most triggers sit on the
      // right edge of the modal header) so the popover does not slide off
      // the right edge of the viewport.
      const right = Math.max(
        POPOVER_VIEWPORT_MARGIN + width,
        Math.min(viewportW - POPOVER_VIEWPORT_MARGIN, rect.right),
      );
      const left = right - width;
      const top = placeAbove
        ? Math.max(POPOVER_VIEWPORT_MARGIN, rect.top - popHeight - POPOVER_VERTICAL_GAP)
        : rect.bottom + POPOVER_VERTICAL_GAP;
      setPos({ top, left, width, placeAbove });
    };
    compute();
    const onResize = () => compute();
    window.addEventListener("scroll", onResize, true);
    window.addEventListener("resize", onResize);
    return () => {
      window.removeEventListener("scroll", onResize, true);
      window.removeEventListener("resize", onResize);
    };
  }, [anchor]);

  useEffect(() => {
    function onDocMouseDown(e: MouseEvent) {
      const target = e.target;
      if (!(target instanceof Node)) return;
      if (popoverRef.current?.contains(target)) return;
      if (anchor?.contains(target)) return;
      onClose();
    }
    document.addEventListener("mousedown", onDocMouseDown);
    return () => document.removeEventListener("mousedown", onDocMouseDown);
  }, [anchor, onClose]);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key !== "Escape") return;
      e.preventDefault();
      e.stopPropagation();
      onClose();
      anchor?.focus();
    }
    window.addEventListener("keydown", onKey, true);
    return () => window.removeEventListener("keydown", onKey, true);
  }, [anchor, onClose]);

  useLayoutEffect(() => {
    if (!pos) return;
    popoverRef.current?.focus();
  }, [pos]);

  if (!anchor) return null;

  return createPortal(
    <div
      ref={popoverRef}
      role="dialog"
      aria-labelledby={titleId}
      aria-modal="false"
      tabIndex={-1}
      data-testid="test-scenarios-popover"
      className="test-scenarios-popover"
      data-place-above={pos?.placeAbove ? "true" : "false"}
      style={{
        position: "fixed",
        top: pos?.top ?? -9999,
        left: pos?.left ?? -9999,
        width: pos?.width ?? POPOVER_DEFAULT_WIDTH,
        visibility: pos ? "visible" : "hidden",
        zIndex: POPOVER_Z_INDEX,
      }}
    >
      <p id={titleId} className="test-scenarios-popover__title">
        Load a sample task
      </p>
      <ul className="test-scenarios-popover__list">
        {TEST_SCENARIOS.map((scenario) => {
          const picked = lastPickedId === scenario.id;
          return (
            <li key={scenario.id}>
              <button
                type="button"
                className="test-scenarios-popover__row"
                data-testid={`test-scenarios-pick-${scenario.id}`}
                onClick={() => {
                  lastPickedScenarioId = scenario.id;
                  setLastPickedId(scenario.id);
                  onPick(scenario);
                }}
              >
                <span className="test-scenarios-popover__marker">
                  {picked ? <CheckGlyph /> : <span className="test-scenarios-popover__bullet" />}
                </span>
                <span className="test-scenarios-popover__row-copy">
                  <span className="test-scenarios-popover__row-title">
                    {scenario.name}
                  </span>
                  <span className="test-scenarios-popover__row-description">
                    {scenario.description}
                  </span>
                </span>
              </button>
            </li>
          );
        })}
      </ul>
    </div>,
    document.body,
  );
}

function CheckGlyph() {
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
      <path d="M20 6 9 17l-5-5" />
    </svg>
  );
}
