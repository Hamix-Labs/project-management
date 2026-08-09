import {
  isChildlessEmptyParagraph,
  type PromptCanvasClickBlock,
  type PromptCanvasClickDecision,
} from "./decidePromptCanvasClick";

/** Minimal BlockNote surface needed to apply a canvas-click decision. */
export type PromptCanvasClickEditor = {
  document: PromptCanvasClickBlock[];
  focus: () => void;
  setTextCursorPosition: (
    block: PromptCanvasClickBlock,
    placement?: "start" | "end",
  ) => void;
  insertBlocks: (
    blocks: Array<{ type: "paragraph" }>,
    reference: PromptCanvasClickBlock,
    placement: "after",
  ) => PromptCanvasClickBlock[];
};

/**
 * Applies a {@link decidePromptCanvasClick} result to the editor.
 *
 * `focusLast` and `focusFirst` touch only the selection. `appendAndFocus`
 * inserts once; callers must only choose it when the trailing block is not
 * already an empty paragraph, so repeat clicks do not dirty autosave.
 */
export function applyPromptCanvasClick(
  editor: PromptCanvasClickEditor,
  decision: PromptCanvasClickDecision,
): void {
  if (decision === "ignore") {
    return;
  }

  const blocks = editor.document;
  if (blocks.length === 0) {
    return;
  }

  if (decision === "focusFirst") {
    editor.setTextCursorPosition(blocks[0]!, "start");
    editor.focus();
    return;
  }

  const last = blocks[blocks.length - 1]!;

  if (decision === "focusLast") {
    editor.setTextCursorPosition(last, "end");
    editor.focus();
    return;
  }

  // Belt-and-suspenders: never append when a reusable slot already exists.
  if (isChildlessEmptyParagraph(last)) {
    editor.setTextCursorPosition(last, "end");
    editor.focus();
    return;
  }

  const inserted = editor.insertBlocks([{ type: "paragraph" }], last, "after")[0];
  if (inserted === undefined) {
    return;
  }
  editor.setTextCursorPosition(inserted, "start");
  editor.focus();
}
