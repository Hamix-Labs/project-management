import { useEffect, type ReactNode } from "react";
import {
  BlockColorsItem,
  DragHandleMenu,
  RemoveBlockItem,
  useDictionary,
} from "@blocknote/react";
import { PromptEditorTurnIntoItem } from "./PromptEditorTurnIntoItem";
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
 * Prompt-owned drag-handle menu. Extends the #158/#166 open-beacon seam with
 * Turn into (#156) between Delete and Colors. Do not fork a second menu —
 * highlight wiring stays in the beacon + provider.
 *
 * Passed to BlockNote's `DragHandleButton` via `dragHandleMenu`.
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
          <PromptEditorTurnIntoItem />
          <BlockColorsItem>{dict.drag_handle.colors_menuitem}</BlockColorsItem>
        </>
      )}
    </DragHandleMenu>
  );
}
