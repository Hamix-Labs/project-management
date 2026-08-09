import { describe, expect, it } from "vitest";
import { decideActiveBlockIds } from "./decideActiveBlockIds";

describe("decideActiveBlockIds", () => {
  it("returns nothing while the side menu is only hovered", () => {
    expect(
      decideActiveBlockIds({
        menuOpen: false,
        dragging: false,
        targetBlockId: "a",
        selectionBlockIds: undefined,
      }),
    ).toEqual([]);
  });

  it("returns nothing without a target block", () => {
    expect(
      decideActiveBlockIds({
        menuOpen: true,
        dragging: false,
        targetBlockId: undefined,
        selectionBlockIds: ["a"],
      }),
    ).toEqual([]);
  });

  it("highlights the handle's block when the drag-handle menu is open", () => {
    expect(
      decideActiveBlockIds({
        menuOpen: true,
        dragging: false,
        targetBlockId: "a",
        selectionBlockIds: undefined,
      }),
    ).toEqual(["a"]);
  });

  it("highlights the handle's block while a drag is in flight", () => {
    expect(
      decideActiveBlockIds({
        menuOpen: false,
        dragging: true,
        targetBlockId: "a",
        selectionBlockIds: undefined,
      }),
    ).toEqual(["a"]);
  });

  it("highlights every selected block when the handle's block is in the selection", () => {
    expect(
      decideActiveBlockIds({
        menuOpen: true,
        dragging: false,
        targetBlockId: "b",
        selectionBlockIds: ["a", "b", "c"],
      }),
    ).toEqual(["a", "b", "c"]);
  });

  it("ignores a selection that does not include the handle's block", () => {
    expect(
      decideActiveBlockIds({
        menuOpen: true,
        dragging: false,
        targetBlockId: "x",
        selectionBlockIds: ["a", "b"],
      }),
    ).toEqual(["x"]);
  });

  it("treats a single-block selection the same as no multi-select", () => {
    expect(
      decideActiveBlockIds({
        menuOpen: true,
        dragging: false,
        targetBlockId: "a",
        selectionBlockIds: ["a"],
      }),
    ).toEqual(["a"]);
  });

  it("keeps the highlight during drag even if the menu is closed", () => {
    expect(
      decideActiveBlockIds({
        menuOpen: false,
        dragging: true,
        targetBlockId: "a",
        selectionBlockIds: ["a", "b"],
      }),
    ).toEqual(["a", "b"]);
  });
});
