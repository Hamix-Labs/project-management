import { useCallback, useRef, type PointerEvent as ReactPointerEvent } from "react";

/** Matches hamix-task-ide `min-h-[320px]` so the brief cannot collapse. */
export const COMPOSE_BRIEF_EDITOR_MIN_PX = 320;

function readEditorHeightPx(root: HTMLElement): number {
  const declared = root.style.getPropertyValue("--compose-brief-editor-h");
  if (declared.endsWith("px")) {
    const parsed = Number.parseFloat(declared);
    if (Number.isFinite(parsed)) return parsed;
  }
  const surface = root.querySelector<HTMLElement>(
    ".tiptap, .ProseMirror, .rich-prompt-editor",
  );
  const measured = surface?.getBoundingClientRect().height ?? 0;
  return Math.max(COMPOSE_BRIEF_EDITOR_MIN_PX, measured);
}

/**
 * Pointer-drag vertical resize for the compose v2 brief editor.
 * Height is stored on `--compose-brief-editor-h` (session only; resets on remount).
 */
export function useComposeBriefVerticalResize() {
  const rootRef = useRef<HTMLElement>(null);
  const dragRef = useRef<{ startY: number; startH: number } | null>(null);

  const onGripPointerDown = useCallback((e: ReactPointerEvent<HTMLElement>) => {
    if (e.button > 0) return;
    const root = rootRef.current;
    if (!root || !Number.isFinite(e.clientY)) return;
    dragRef.current = {
      startY: e.clientY,
      startH: readEditorHeightPx(root),
    };
    root.dataset.resizing = "true";
    if (typeof e.currentTarget.setPointerCapture === "function") {
      e.currentTarget.setPointerCapture(e.pointerId);
    }
    e.preventDefault();
  }, []);

  const onGripPointerMove = useCallback((e: ReactPointerEvent<HTMLElement>) => {
    const drag = dragRef.current;
    const root = rootRef.current;
    if (!drag || !root || !Number.isFinite(e.clientY)) return;
    const next = Math.max(
      COMPOSE_BRIEF_EDITOR_MIN_PX,
      drag.startH + (e.clientY - drag.startY),
    );
    root.style.setProperty("--compose-brief-editor-h", `${next}px`);
  }, []);

  const endDrag = useCallback((e: ReactPointerEvent<HTMLElement>) => {
    if (!dragRef.current) return;
    dragRef.current = null;
    if (rootRef.current) delete rootRef.current.dataset.resizing;
    const target = e.currentTarget;
    if (
      typeof target.hasPointerCapture === "function" &&
      e.pointerId != null &&
      target.hasPointerCapture(e.pointerId)
    ) {
      target.releasePointerCapture(e.pointerId);
    }
  }, []);

  return {
    rootRef,
    onGripPointerDown,
    onGripPointerMove,
    onGripPointerUp: endDrag,
    onGripPointerCancel: endDrag,
  };
}
