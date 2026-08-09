import { useEffect, type ReactNode } from "react";
import {
  BlockColorsItem,
  DragHandleMenu,
  RemoveBlockItem,
  useDictionary,
} from "@blocknote/react";
import { usePromptDragHandleMenuOpenChange } from "./promptDragHandleMenuOpenContext";

/**
 * Mounts with Ariakit's drag-handle dropdown (`unmountOnHide`) and reports
 * open/close to {@link PromptDragHandleMenuOpenProvider}. Renders nothing.
 */
function DragHandleMenuOpenBeacon() {
  const onOpenChange = usePromptDragHandleMenuOpenChange();

  useEffect(() => {
    onOpenChange?.(true);
    return () => onOpenChange?.(false);
  }, [onOpenChange]);

  return null;
}

/**
 * Prompt-owned drag-handle menu. Stock contents today (Delete, Colors); sibling
 * issue #156 adds "Turn into" here without touching the highlight wiring.
 *
 * Passed to BlockNote's `DragHandleButton` via `dragHandleMenu`, so we observe
 * menu-open without forking the button (and without binding highlight to the
 * side menu's hover `show` flag).
 */
export function PromptEditorDragHandleMenu({
  children,
}: {
  children?: ReactNode;
}) {
  const dict = useDictionary();

  return (
    <DragHandleMenu>
      <DragHandleMenuOpenBeacon />
      {children ?? (
        <>
          <RemoveBlockItem>{dict.drag_handle.delete_menuitem}</RemoveBlockItem>
          <BlockColorsItem>{dict.drag_handle.colors_menuitem}</BlockColorsItem>
        </>
      )}
    </DragHandleMenu>
  );
}
