import { flushSync } from "react-dom";

function prefersReducedMotion(): boolean {
  return (
    typeof window !== "undefined" &&
    typeof window.matchMedia === "function" &&
    window.matchMedia("(prefers-reduced-motion: reduce)").matches
  );
}

/**
 * Runs `update` inside `document.startViewTransition` when available and
 * motion is allowed. Returns whether a view transition was started.
 */
export function runViewTransition(update: () => void): boolean {
  const doc = document as Document & {
    startViewTransition?: (cb: () => void) => unknown;
  };
  if (
    prefersReducedMotion() ||
    typeof doc.startViewTransition !== "function"
  ) {
    return false;
  }
  doc.startViewTransition(() => {
    flushSync(update);
  });
  return true;
}
