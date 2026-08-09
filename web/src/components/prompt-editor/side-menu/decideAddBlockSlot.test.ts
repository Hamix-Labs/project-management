import { describe, expect, it } from "vitest";
import { decideAddBlockSlot } from "./decideAddBlockSlot";

const emptyParagraph = { type: "paragraph", content: [] };
const filledParagraph = {
  type: "paragraph",
  content: [{ type: "text", text: "brief" }],
};

describe("decideAddBlockSlot", () => {
  it("is a no-op when no block is hovered", () => {
    expect(decideAddBlockSlot(undefined, emptyParagraph)).toBeUndefined();
  });

  it("focuses the hovered block when it is already empty", () => {
    expect(decideAddBlockSlot(emptyParagraph, filledParagraph)).toBe(
      "focusHovered",
    );
  });

  it("reuses an empty paragraph that already follows the hovered block", () => {
    expect(decideAddBlockSlot(filledParagraph, emptyParagraph)).toBe(
      "focusNext",
    );
  });

  it("inserts when the hovered block is last", () => {
    expect(decideAddBlockSlot(filledParagraph, undefined)).toBe("insertAfter");
  });

  it("inserts when the next block has content", () => {
    expect(decideAddBlockSlot(filledParagraph, filledParagraph)).toBe(
      "insertAfter",
    );
  });

  it("inserts rather than hijacking an empty heading below", () => {
    expect(
      decideAddBlockSlot(filledParagraph, { type: "heading", content: [] }),
    ).toBe("insertAfter");
  });

  it("inserts rather than hijacking an empty list item below", () => {
    expect(
      decideAddBlockSlot(filledParagraph, {
        type: "bulletListItem",
        content: [],
      }),
    ).toBe("insertAfter");
  });

  it("inserts when the empty paragraph below nests children", () => {
    expect(
      decideAddBlockSlot(filledParagraph, {
        type: "paragraph",
        content: [],
        children: [filledParagraph],
      }),
    ).toBe("insertAfter");
  });

  it("treats contentless blocks as not empty in either position", () => {
    const embed = { type: "repoFileEmbed" };

    expect(decideAddBlockSlot(embed, filledParagraph)).toBe("insertAfter");
    expect(decideAddBlockSlot(filledParagraph, embed)).toBe("insertAfter");
  });

  it("stays idempotent once the slot below exists", () => {
    const first = decideAddBlockSlot(filledParagraph, undefined);
    const second = decideAddBlockSlot(filledParagraph, emptyParagraph);
    const third = decideAddBlockSlot(filledParagraph, emptyParagraph);

    expect(first).toBe("insertAfter");
    expect(second).toBe("focusNext");
    expect(third).toBe("focusNext");
  });
});
