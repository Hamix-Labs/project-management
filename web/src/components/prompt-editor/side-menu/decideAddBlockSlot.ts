/** Structural subset of a BlockNote block, so this policy stays schema-agnostic. */
export type AddBlockSlotBlock = {
  type: string;
  content?: unknown;
  children?: unknown[];
};

export type AddBlockSlot = "focusHovered" | "focusNext" | "insertAfter";

/** Mirrors BlockNote's own emptiness test: `content: "none"` blocks are not empty, they are contentless. */
function hasEmptyContent(block: AddBlockSlotBlock): boolean {
  return Array.isArray(block.content) && block.content.length === 0;
}

function isReusableSlot(block: AddBlockSlotBlock | undefined): boolean {
  if (block === undefined || !hasEmptyContent(block)) {
    return false;
  }

  // Only paragraphs: the stock insert produces a paragraph, so reusing an empty
  // heading or list item below would silently hand the user a different block
  // type than the one they asked for.
  if (block.type !== "paragraph") {
    return false;
  }

  // An empty paragraph with children is a container for the blocks nested under
  // it, not a free slot — typing there writes at a different indentation level.
  return (block.children?.length ?? 0) === 0;
}

/**
 * Resolves what the side menu's add-block affordance should do for the hovered
 * block, given the block that follows it.
 *
 * BlockNote's stock button always inserts a paragraph after a non-empty block,
 * so clicking it, dismissing the slash menu, and clicking again stacks a second
 * empty paragraph — and dirties the document into an autosave each time. The
 * click is an intent to reach the writing slot below the block, not to create a
 * block, so an already-empty slot is reused instead.
 *
 * `next` must be the hovered block's next *sibling* (`editor.getNextBlock`),
 * which is exactly where an insert would land, making reuse its precise inverse.
 *
 * @returns `undefined` when there is no hovered block and the click is a no-op.
 */
export function decideAddBlockSlot(
  hovered: AddBlockSlotBlock | undefined,
  next: AddBlockSlotBlock | undefined,
): AddBlockSlot | undefined {
  if (hovered === undefined) {
    return undefined;
  }

  if (hasEmptyContent(hovered)) {
    return "focusHovered";
  }

  return isReusableSlot(next) ? "focusNext" : "insertAfter";
}
