/** Structural subset of a BlockNote block for emptiness checks. */
export type PromptCanvasClickBlock = {
  type: string;
  content?: unknown;
  children?: unknown[];
};

/**
 * Where a pointer event landed relative to document chrome and block content.
 *
 * Geometry is classified by the adapter; this module only decides what to do.
 */
export type PromptCanvasClickRegion =
  | "headerChrome"
  | "aboveFirstBlock"
  | "onBlock"
  | "belowLastBlock"
  | "other";

export type PromptCanvasClickDecision =
  | "ignore"
  | "focusFirst"
  | "focusLast"
  | "appendAndFocus";

export type PromptCanvasClickInput = {
  region: PromptCanvasClickRegion;
  /** Controls we must not steal (links, buttons, title input, pickers, …). */
  hitInteractive: boolean;
  /** Click-drag that selected text — must not hijack into a focus call. */
  isSelectingText: boolean;
  hasBlocks: boolean;
  /**
   * Last top-level block is a childless empty paragraph. Reuse it instead of
   * inserting — stray appends would fire onChange and autosave junk HTML (#153).
   */
  lastBlockIsEmptyParagraph: boolean;
};

export type PromptCanvasClickGeometry = {
  clientY: number;
  /** Bottom edge of the meta/divider row (gap below this is aboveFirstBlock). */
  headerBottom: number;
  firstBlockTop: number | null;
  lastBlockBottom: number | null;
  targetInHeaderChrome: boolean;
  targetInBlock: boolean;
};

/**
 * True when the block is a reusable trailing writing slot: a childless empty
 * paragraph. Mirrors the add-block slot policy (ADR-0098) so canvas clicks and
 * `+` agree on what counts as an empty slot.
 */
export function isChildlessEmptyParagraph(
  block: PromptCanvasClickBlock | undefined,
): boolean {
  if (block === undefined || block.type !== "paragraph") {
    return false;
  }
  if (!Array.isArray(block.content) || block.content.length !== 0) {
    return false;
  }
  return (block.children?.length ?? 0) === 0;
}

/**
 * Classifies a click from measured geometry and target ancestry.
 *
 * Pure so unit tests can cover zones without mounting BlockNote.
 */
export function resolvePromptCanvasClickRegion(
  geometry: PromptCanvasClickGeometry,
): PromptCanvasClickRegion {
  if (geometry.targetInHeaderChrome) {
    return "headerChrome";
  }
  if (geometry.targetInBlock) {
    return "onBlock";
  }

  if (
    geometry.firstBlockTop != null &&
    geometry.clientY >= geometry.headerBottom &&
    geometry.clientY < geometry.firstBlockTop
  ) {
    return "aboveFirstBlock";
  }

  if (
    geometry.lastBlockBottom != null &&
    geometry.clientY > geometry.lastBlockBottom
  ) {
    return "belowLastBlock";
  }

  return "other";
}

/**
 * Decides how a canvas click should move the caret (Notion-style empty space).
 *
 * - Gap between divider and first block → caret in the first block.
 * - Empty area below the last block → caret at end of last block; append an
 *   empty paragraph only when the last block is not already one.
 * - Interactive targets, text-selection drags, header chrome, and in-block
 *   clicks are no-ops so we never steal real editing gestures.
 */
export function decidePromptCanvasClick(
  input: PromptCanvasClickInput,
): PromptCanvasClickDecision {
  if (input.hitInteractive || input.isSelectingText || !input.hasBlocks) {
    return "ignore";
  }

  switch (input.region) {
    case "aboveFirstBlock":
      return "focusFirst";
    case "belowLastBlock":
      return input.lastBlockIsEmptyParagraph ? "focusLast" : "appendAndFocus";
    default:
      return "ignore";
  }
}
