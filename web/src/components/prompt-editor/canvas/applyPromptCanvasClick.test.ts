import { describe, expect, it, vi } from "vitest";
import { applyPromptCanvasClick } from "./applyPromptCanvasClick";
import type { PromptCanvasClickEditor } from "./applyPromptCanvasClick";

function mockEditor(
  document: PromptCanvasClickEditor["document"],
): PromptCanvasClickEditor & {
  focus: ReturnType<typeof vi.fn>;
  setTextCursorPosition: ReturnType<typeof vi.fn>;
  insertBlocks: ReturnType<typeof vi.fn>;
} {
  const focus = vi.fn();
  const setTextCursorPosition = vi.fn();
  const insertBlocks = vi.fn(
    (
      _blocks: Array<{ type: "paragraph" }>,
      _reference: unknown,
      _placement: "after",
    ) => [{ type: "paragraph", content: [] }],
  );

  return {
    document,
    focus,
    setTextCursorPosition,
    insertBlocks,
  };
}

describe("applyPromptCanvasClick", () => {
  const first = {
    type: "paragraph",
    content: [{ type: "text", text: "alpha" }],
  };
  const lastFilled = {
    type: "paragraph",
    content: [{ type: "text", text: "omega" }],
  };
  const lastEmpty = { type: "paragraph", content: [] };

  it("is a no-op for ignore", () => {
    const editor = mockEditor([first]);
    applyPromptCanvasClick(editor, "ignore");
    expect(editor.focus).not.toHaveBeenCalled();
    expect(editor.insertBlocks).not.toHaveBeenCalled();
  });

  it("focuses the start of the first block", () => {
    const editor = mockEditor([first, lastFilled]);
    applyPromptCanvasClick(editor, "focusFirst");
    expect(editor.setTextCursorPosition).toHaveBeenCalledWith(first, "start");
    expect(editor.focus).toHaveBeenCalledOnce();
    expect(editor.insertBlocks).not.toHaveBeenCalled();
  });

  it("focuses the end of the last block without inserting", () => {
    const editor = mockEditor([first, lastEmpty]);
    applyPromptCanvasClick(editor, "focusLast");
    expect(editor.setTextCursorPosition).toHaveBeenCalledWith(lastEmpty, "end");
    expect(editor.insertBlocks).not.toHaveBeenCalled();
  });

  it("appends a paragraph after a non-empty last block", () => {
    const editor = mockEditor([first, lastFilled]);
    applyPromptCanvasClick(editor, "appendAndFocus");
    expect(editor.insertBlocks).toHaveBeenCalledWith(
      [{ type: "paragraph" }],
      lastFilled,
      "after",
    );
    expect(editor.setTextCursorPosition).toHaveBeenCalledWith(
      { type: "paragraph", content: [] },
      "start",
    );
  });

  it("refuses to append when the last block is already an empty paragraph", () => {
    const editor = mockEditor([first, lastEmpty]);
    applyPromptCanvasClick(editor, "appendAndFocus");
    expect(editor.insertBlocks).not.toHaveBeenCalled();
    expect(editor.setTextCursorPosition).toHaveBeenCalledWith(lastEmpty, "end");
  });
});
