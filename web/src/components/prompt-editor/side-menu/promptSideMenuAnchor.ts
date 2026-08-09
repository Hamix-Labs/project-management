import { findPromptBlockElement } from "./promptBlockElement";

const COLUMN_SELECTOR = '[data-node-type="column"]';

/**
 * Resolves the side menu's anchor rect for a block id against the live DOM.
 *
 * Must be called on every position pass rather than captured once: ProseMirror
 * replaces a block's DOM node when the block moves (drag & drop, undo, a delete
 * above it), and a captured node keeps reporting its pre-move rect once it is
 * detached. Deriving the rect from the id instead keeps the menu correct without
 * needing a React re-render or a remount.
 *
 * Geometry matches BlockNote's own side menu plugin: x comes from the enclosing
 * column (or the root block group) so the buttons sit in the gutter rather than
 * over the text, while y/width/height come from the block container.
 */
export function promptSideMenuAnchorRect(
  editorDom: HTMLElement | null | undefined,
  blockId: string,
): DOMRect | undefined {
  const blockGroup = editorDom?.firstElementChild;
  if (!editorDom || !blockGroup) {
    return undefined;
  }

  const block = findPromptBlockElement(editorDom, blockId);
  if (!block) {
    return undefined;
  }

  // Columns carry their own padding, so the gutter is measured from the
  // column's first block rather than the column itself.
  const gutterSource =
    block.closest(COLUMN_SELECTOR)?.firstElementChild ?? blockGroup;
  const blockBox = block.getBoundingClientRect();

  return new DOMRect(
    gutterSource.getBoundingClientRect().x,
    blockBox.y,
    blockBox.width,
    blockBox.height,
  );
}
