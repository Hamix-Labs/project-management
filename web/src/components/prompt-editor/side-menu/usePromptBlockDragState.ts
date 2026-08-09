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
 * `drop` and `dragend` are both listened for in the bubble phase, after
 * ProseMirror's own capture-phase handlers have applied the resulting
 * transaction — so by the time this flips back to `false` the moved block is
 * already at its new position and the side menu can be re-measured there.
 * Neither event alone is reliable: dropping on an invalid target fires only
 * `dragend`, and BlockNote has historically seen drops without a matching
 * `dragend` when dragging across editors.
 */
export function usePromptBlockDragState(editorHost: HTMLElement | null) {
  const [dragging, setDragging] = useState(false);

  useEffect(() => {
    if (!editorHost) {
      return;
    }

    const { ownerDocument } = editorHost;

    const onDragStart = (event: DragEvent) => {
      if (isPromptEditorDrag(event, editorHost)) {
        setDragging(true);
      }
    };
    const onDragFinish = () => setDragging(false);

    ownerDocument.addEventListener("dragstart", onDragStart);
    ownerDocument.addEventListener("drop", onDragFinish);
    ownerDocument.addEventListener("dragend", onDragFinish);

    return () => {
      ownerDocument.removeEventListener("dragstart", onDragStart);
      ownerDocument.removeEventListener("drop", onDragFinish);
      ownerDocument.removeEventListener("dragend", onDragFinish);
      setDragging(false);
    };
  }, [editorHost]);

  return dragging;
}
