/**
 * Inputs for deciding which prompt-editor blocks show the active-block highlight.
 *
 * `menuOpen` is the drag-handle *menu* (Delete / Colors / …), not the side
 * menu's hover `show` flag — those are distinct, and only the former should
 * light a block up.
 */
export type ActiveBlockHighlightInput = {
  menuOpen: boolean;
  dragging: boolean;
  targetBlockId: string | undefined;
  /**
   * Ids from `editor.getSelection()?.blocks` when a selection exists.
   * Pass `undefined` when there is no selection (or it has not been read yet).
   */
  selectionBlockIds: readonly string[] | undefined;
};

/**
 * Which block ids should carry the active-block highlight.
 *
 * Highlight is on while the drag-handle menu is open **or** a block drag is in
 * flight — never on mere side-menu hover. When the handle's block is part of a
 * multi-block selection, every selected block is returned so the visual matches
 * what BlockNote's RemoveBlockItem will delete.
 */
export function decideActiveBlockIds(
  input: ActiveBlockHighlightInput,
): string[] {
  const { menuOpen, dragging, targetBlockId, selectionBlockIds } = input;

  if (targetBlockId === undefined) {
    return [];
  }

  if (!menuOpen && !dragging) {
    return [];
  }

  if (
    selectionBlockIds !== undefined &&
    selectionBlockIds.includes(targetBlockId)
  ) {
    return [...selectionBlockIds];
  }

  return [targetBlockId];
}
