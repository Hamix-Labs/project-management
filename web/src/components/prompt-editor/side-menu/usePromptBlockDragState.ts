import { useEffect, useState } from "react";

const BLOCKNOTE_DRAG_TYPE = "blocknote/html";

/**
 * A drag belongs to this editor if BlockNote put its block payload on the
 * clipboard, or if the drag simply started somewhere inside the editor host
 * (plain text selection drags carry no BlockNote payload).
 */
export function isPromptEditorDrag(
  event: DragEvent,
  editorHost: HTMLElement,
): boolean {
  if (
    Array.from(event.dataTransfer?.types ?? []).includes(BLOCKNOTE_DRAG_TYPE)
  ) {
    return true;
  }

  return event.target instanceof Node && editorHost.contains(event.target);
}

/**
 * Tracks whether a drag started inside the prompt editor is still in flight.
 *
 * The flag is raised in a task queued *after* `dragstart` rather than during
 * it. Chrome abandons a drag whose source element is restyled from within the
 * dragstart handler, and the drag handle sits inside the menu that this flag
 * hides — flipping it synchronously kills the drag before it begins. A timeout
 * rather than an animation frame guarantees the browser has finished
 * initiating the drag, and the delay is far too short to be visible.
 *
 * It is lowered on bubble-phase `drop` and `dragend`, after ProseMirror's
 * capture-phase handlers have applied the resulting transaction, so the moved
 * block is already in its new position when the menu is re-measured. Neither
 * event alone is reliable: dropping on an invalid target fires only `dragend`,
 * and BlockNote has historically seen drops without a matching `dragend` when
 * dragging between editors. `mousemove` is the final backstop — the browser
 * does not fire it while a drag is in flight, so seeing one proves there isn't.
 */
export function usePromptBlockDragState(editorHost: HTMLElement | null) {
  const [dragging, setDragging] = useState(false);

  useEffect(() => {
    if (!editorHost) {
      return;
    }

    const { ownerDocument } = editorHost;
    const view = ownerDocument.defaultView ?? window;
    let pending = 0;
    let active = false;

    const stopDragging = () => {
      if (!active && !pending) {
        return;
      }
      if (pending) {
        view.clearTimeout(pending);
        pending = 0;
      }
      active = false;
      setDragging(false);
    };

    const onDragStart = (event: DragEvent) => {
      if (pending || active || !isPromptEditorDrag(event, editorHost)) {
        return;
      }
      pending = view.setTimeout(() => {
        pending = 0;
        active = true;
        setDragging(true);
      }, 0);
    };

    ownerDocument.addEventListener("dragstart", onDragStart);
    ownerDocument.addEventListener("drop", stopDragging);
    ownerDocument.addEventListener("dragend", stopDragging);
    ownerDocument.addEventListener("mousemove", stopDragging, {
      passive: true,
    });

    return () => {
      if (pending) {
        view.clearTimeout(pending);
      }
      ownerDocument.removeEventListener("dragstart", onDragStart);
      ownerDocument.removeEventListener("drop", stopDragging);
      ownerDocument.removeEventListener("dragend", stopDragging);
      ownerDocument.removeEventListener("mousemove", stopDragging);
    };
  }, [editorHost]);

  return dragging;
}
