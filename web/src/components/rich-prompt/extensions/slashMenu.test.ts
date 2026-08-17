// @vitest-environment jsdom
import { Editor } from "@tiptap/core";
import StarterKit from "@tiptap/starter-kit";
import { describe, expect, it, vi } from "vitest";
import {
  runSlashCommand,
  SLASH_ITEMS,
  slashItemsForCatalog,
  SlashMenu,
} from "./slashMenu";
import { filterSlashItems } from "./slashMenuList";

function makeEditor(onAiTrigger: (msg: string) => void = () => {}) {
  return new Editor({
    extensions: [StarterKit, SlashMenu.configure({ onAiTrigger })],
    content: "<p></p>",
  });
}

describe("SLASH_ITEMS", () => {
  it("includes heading, list, quote, mention and ai commands", () => {
    const ids = SLASH_ITEMS.map((i) => i.id);
    expect(ids).toEqual(
      expect.arrayContaining([
        "heading-2",
        "heading-3",
        "bullet-list",
        "ordered-list",
        "blockquote",
        "mention",
        "ai",
      ]),
    );
  });
});

describe("slashItemsForCatalog", () => {
  it("returns the full catalog for all", () => {
    expect(slashItemsForCatalog("all")).toEqual(SLASH_ITEMS);
  });

  it("keeps only Hamix product commands", () => {
    expect(slashItemsForCatalog("commands").map((i) => i.id)).toEqual([
      "mention",
      "ai",
    ]);
  });
});

describe("filterSlashItems", () => {
  it("returns everything for an empty query", () => {
    expect(filterSlashItems(SLASH_ITEMS, "")).toEqual(SLASH_ITEMS);
  });

  it("matches on id, label, and keywords", () => {
    const byHeading = filterSlashItems(SLASH_ITEMS, "heading").map((i) => i.id);
    expect(byHeading).toEqual(
      expect.arrayContaining(["heading-2", "heading-3"]),
    );
    const byBullet = filterSlashItems(SLASH_ITEMS, "bullet").map((i) => i.id);
    expect(byBullet).toEqual(["bullet-list"]);
    const byAi = filterSlashItems(SLASH_ITEMS, "ai").map((i) => i.id);
    expect(byAi).toEqual(["ai"]);
  });

  it("returns an empty list when nothing matches", () => {
    expect(filterSlashItems(SLASH_ITEMS, "zzzz")).toEqual([]);
  });
});

describe("runSlashCommand", () => {
  it("converts the current block to a heading 2", () => {
    const editor = makeEditor();
    editor.commands.focus("start");
    const range = { from: 1, to: 1 };
    runSlashCommand(editor, range, "heading-2");
    expect(editor.getHTML()).toContain("<h2");
    editor.destroy();
  });

  it("toggles a bullet list", () => {
    const editor = makeEditor();
    editor.commands.focus("start");
    runSlashCommand(editor, { from: 1, to: 1 }, "bullet-list");
    expect(editor.getHTML()).toContain("<ul");
    editor.destroy();
  });

  it("inserts an @ character when mention is chosen", () => {
    const editor = makeEditor();
    editor.commands.focus("start");
    runSlashCommand(editor, { from: 1, to: 1 }, "mention");
    expect(editor.getText()).toBe("@");
    editor.destroy();
  });

  it("invokes onAiTrigger for the ai command and leaves no residual text", () => {
    const onAiTrigger = vi.fn();
    const editor = makeEditor(onAiTrigger);
    editor.commands.focus("start");
    runSlashCommand(editor, { from: 1, to: 1 }, "ai", onAiTrigger);
    expect(onAiTrigger).toHaveBeenCalledWith("");
    expect(editor.getText()).toBe("");
    editor.destroy();
  });
});
