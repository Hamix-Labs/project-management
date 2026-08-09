// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { promptSideMenuAnchorRect } from "./promptSideMenuAnchor";

type Rect = { x: number; y: number; width: number; height: number };

function stubRect(element: Element, rect: Rect) {
  element.getBoundingClientRect = () => DOMRect.fromRect(rect);
}

function buildEditorDom() {
  const editorDom = document.createElement("div");
  const blockGroup = document.createElement("div");
  blockGroup.className = "bn-block-group";
  stubRect(blockGroup, { x: 40, y: 0, width: 600, height: 400 });
  editorDom.appendChild(blockGroup);
  return { editorDom, blockGroup };
}

function appendBlock(parent: Element, blockId: string, rect: Rect) {
  const outer = document.createElement("div");
  outer.className = "bn-block-outer";
  outer.setAttribute("data-node-type", "blockOuter");
  outer.setAttribute("data-id", blockId);
  stubRect(outer, { ...rect, x: rect.x - 4 });

  const container = document.createElement("div");
  container.className = "bn-block";
  container.setAttribute("data-node-type", "blockContainer");
  container.setAttribute("data-id", blockId);
  stubRect(container, rect);

  outer.appendChild(container);
  parent.appendChild(outer);
  return { outer, container };
}

describe("promptSideMenuAnchorRect", () => {
  it("takes the gutter x from the block group and the box from the block", () => {
    const { editorDom, blockGroup } = buildEditorDom();
    appendBlock(blockGroup, "block-1", { x: 60, y: 120, width: 560, height: 28 });

    const rect = promptSideMenuAnchorRect(editorDom, "block-1");

    expect(rect).toBeDefined();
    expect(rect?.x).toBe(40);
    expect(rect?.y).toBe(120);
    expect(rect?.width).toBe(560);
    expect(rect?.height).toBe(28);
  });

  it("follows the block when ProseMirror replaces its node after a move", () => {
    const { editorDom, blockGroup } = buildEditorDom();
    const first = appendBlock(blockGroup, "block-1", {
      x: 60,
      y: 120,
      width: 560,
      height: 28,
    });
    appendBlock(blockGroup, "block-2", { x: 60, y: 160, width: 560, height: 28 });

    expect(promptSideMenuAnchorRect(editorDom, "block-1")?.y).toBe(120);

    // A drop re-renders the moved block as a brand new node carrying the same
    // id. Anything holding the old node keeps reporting the pre-move rect.
    first.outer.remove();
    appendBlock(blockGroup, "block-1", { x: 60, y: 200, width: 560, height: 28 });

    expect(promptSideMenuAnchorRect(editorDom, "block-1")?.y).toBe(200);
  });

  it("measures the gutter from the first block of an enclosing column", () => {
    const { editorDom, blockGroup } = buildEditorDom();
    const column = document.createElement("div");
    column.setAttribute("data-node-type", "column");
    stubRect(column, { x: 300, y: 100, width: 300, height: 200 });
    blockGroup.appendChild(column);

    const firstInColumn = appendBlock(column, "block-1", {
      x: 316,
      y: 108,
      width: 268,
      height: 28,
    });
    stubRect(firstInColumn.outer, { x: 312, y: 108, width: 272, height: 28 });
    appendBlock(column, "block-2", { x: 316, y: 148, width: 268, height: 28 });

    expect(promptSideMenuAnchorRect(editorDom, "block-2")?.x).toBe(312);
  });

  it("returns undefined when the block or the block group is missing", () => {
    const { editorDom, blockGroup } = buildEditorDom();
    appendBlock(blockGroup, "block-1", { x: 60, y: 120, width: 560, height: 28 });

    expect(promptSideMenuAnchorRect(editorDom, "block-missing")).toBeUndefined();
    expect(promptSideMenuAnchorRect(null, "block-1")).toBeUndefined();
    expect(
      promptSideMenuAnchorRect(document.createElement("div"), "block-1"),
    ).toBeUndefined();
  });
});
