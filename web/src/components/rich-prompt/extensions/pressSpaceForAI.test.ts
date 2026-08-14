// @vitest-environment jsdom
import { Editor } from "@tiptap/core";
import StarterKit from "@tiptap/starter-kit";
import { describe, expect, it, vi } from "vitest";
import {
  isCaretAtEmptyBlockStart,
  PressSpaceForAI,
} from "./pressSpaceForAI";

function makeEditor(onAiTrigger: (msg: string) => void, content = "<p></p>") {
  return new Editor({
    extensions: [StarterKit, PressSpaceForAI.configure({ onAiTrigger })],
    content,
  });
}

/** Simulate a keydown event on the editor's contenteditable node. */
function dispatchKey(editor: Editor, key: string): boolean {
  const event = new KeyboardEvent("keydown", {
    key,
    bubbles: true,
    cancelable: true,
  });
  editor.view.dom.dispatchEvent(event);
  return event.defaultPrevented;
}

describe("isCaretAtEmptyBlockStart", () => {
  it("returns true for an empty paragraph with the caret at offset 0", () => {
    const editor = makeEditor(vi.fn());
    editor.commands.focus("start");
    expect(isCaretAtEmptyBlockStart(editor.state)).toBe(true);
    editor.destroy();
  });

  it("returns false when the paragraph has any text", () => {
    const editor = makeEditor(vi.fn(), "<p>hi</p>");
    editor.commands.focus("start");
    expect(isCaretAtEmptyBlockStart(editor.state)).toBe(false);
    editor.destroy();
  });

  it("returns false when the caret is not at the block start", () => {
    const editor = makeEditor(vi.fn(), "<p>hello</p>");
    editor.commands.focus("end");
    expect(isCaretAtEmptyBlockStart(editor.state)).toBe(false);
    editor.destroy();
  });
});

describe("PressSpaceForAI", () => {
  it("fires onAiTrigger and prevents the space when the caret is at the empty block start", () => {
    const onAiTrigger = vi.fn();
    const editor = makeEditor(onAiTrigger);
    editor.commands.focus("start");

    const prevented = dispatchKey(editor, " ");

    expect(prevented).toBe(true);
    expect(onAiTrigger).toHaveBeenCalledWith("");
    expect(editor.getText()).toBe("");
    editor.destroy();
  });

  it("lets space fall through when the block already has text", () => {
    const onAiTrigger = vi.fn();
    const editor = makeEditor(onAiTrigger, "<p>hello</p>");
    editor.commands.focus("end");

    const prevented = dispatchKey(editor, " ");

    expect(prevented).toBe(false);
    expect(onAiTrigger).not.toHaveBeenCalled();
    editor.destroy();
  });

  it("does nothing when onAiTrigger is not provided", () => {
    const editor = new Editor({
      extensions: [StarterKit, PressSpaceForAI],
      content: "<p></p>",
    });
    editor.commands.focus("start");

    expect(() => dispatchKey(editor, " ")).not.toThrow();
    editor.destroy();
  });
});
