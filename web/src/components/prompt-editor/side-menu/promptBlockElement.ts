/** Matches BlockNote's block container node attribute. */
export const PROMPT_BLOCK_CONTAINER_SELECTOR =
  '[data-node-type="blockContainer"]';

/**
 * Attribute stamped on a block container while it is the drag-handle menu
 * target (or part of the multi-block selection the menu will act on) / while
 * that block is being dragged. CSS keys off this — nothing sets a ProseMirror
 * NodeSelection.
 */
export const PROMPT_BLOCK_ACTIVE_ATTR = "data-prompt-block-active";

/**
 * Resolves the live block-container element for a block id.
 *
 * Must be re-queried rather than cached: ProseMirror replaces the DOM node when
 * a block moves, and a captured element keeps its pre-move identity once
 * detached.
 */
export function findPromptBlockElement(
  editorDom: HTMLElement | null | undefined,
  blockId: string,
): Element | null {
  if (!editorDom) {
    return null;
  }

  return editorDom.querySelector(
    `${PROMPT_BLOCK_CONTAINER_SELECTOR}[data-id="${CSS.escape(blockId)}"]`,
  );
}
