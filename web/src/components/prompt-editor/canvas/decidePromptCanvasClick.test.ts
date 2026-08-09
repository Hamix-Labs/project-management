import { describe, expect, it } from "vitest";
import {
  decidePromptCanvasClick,
  isChildlessEmptyParagraph,
  resolvePromptCanvasClickRegion,
  type PromptCanvasClickGeometry,
  type PromptCanvasClickInput,
} from "./decidePromptCanvasClick";

const emptyParagraph = { type: "paragraph", content: [] };
const filledParagraph = {
  type: "paragraph",
  content: [{ type: "text", text: "brief" }],
};

function decide(
  overrides: Partial<PromptCanvasClickInput>,
): ReturnType<typeof decidePromptCanvasClick> {
  return decidePromptCanvasClick({
    region: "belowLastBlock",
    hitInteractive: false,
    isSelectingText: false,
    hasBlocks: true,
    lastBlockIsEmptyParagraph: false,
    ...overrides,
  });
}

function region(
  overrides: Partial<PromptCanvasClickGeometry>,
): ReturnType<typeof resolvePromptCanvasClickRegion> {
  return resolvePromptCanvasClickRegion({
    clientY: 200,
    headerBottom: 100,
    firstBlockTop: 140,
    lastBlockBottom: 400,
    targetInHeaderChrome: false,
    targetInBlock: false,
    ...overrides,
  });
}

describe("isChildlessEmptyParagraph", () => {
  it("accepts a childless empty paragraph", () => {
    expect(isChildlessEmptyParagraph(emptyParagraph)).toBe(true);
  });

  it("rejects filled paragraphs, headings, and nested empty paragraphs", () => {
    expect(isChildlessEmptyParagraph(filledParagraph)).toBe(false);
    expect(
      isChildlessEmptyParagraph({ type: "heading", content: [] }),
    ).toBe(false);
    expect(
      isChildlessEmptyParagraph({
        type: "paragraph",
        content: [],
        children: [filledParagraph],
      }),
    ).toBe(false);
    expect(isChildlessEmptyParagraph(undefined)).toBe(false);
  });
});

describe("resolvePromptCanvasClickRegion", () => {
  it("classifies the gap between the divider and the first block", () => {
    expect(region({ clientY: 120 })).toBe("aboveFirstBlock");
  });

  it("classifies empty space below the last block", () => {
    expect(region({ clientY: 450 })).toBe("belowLastBlock");
  });

  it("defers to header chrome and in-block targets", () => {
    expect(region({ targetInHeaderChrome: true, clientY: 120 })).toBe(
      "headerChrome",
    );
    expect(region({ targetInBlock: true, clientY: 450 })).toBe("onBlock");
  });

  it("ignores clicks above the divider or between block tops and bottoms", () => {
    expect(region({ clientY: 80 })).toBe("other");
    expect(region({ clientY: 250 })).toBe("other");
  });

  it("treats a missing first block top as not above-first", () => {
    expect(region({ firstBlockTop: null, clientY: 120 })).toBe("other");
  });
});

describe("decidePromptCanvasClick", () => {
  it("focuses the first block for clicks above it", () => {
    expect(decide({ region: "aboveFirstBlock" })).toBe("focusFirst");
  });

  it("appends when the last block is not an empty paragraph", () => {
    expect(decide({ region: "belowLastBlock" })).toBe("appendAndFocus");
  });

  it("reuses an empty trailing paragraph without appending", () => {
    expect(
      decide({
        region: "belowLastBlock",
        lastBlockIsEmptyParagraph: true,
      }),
    ).toBe("focusLast");
  });

  it("ignores interactive targets", () => {
    expect(
      decide({ region: "aboveFirstBlock", hitInteractive: true }),
    ).toBe("ignore");
    expect(
      decide({ region: "belowLastBlock", hitInteractive: true }),
    ).toBe("ignore");
  });

  it("ignores text-selection drags", () => {
    expect(
      decide({ region: "aboveFirstBlock", isSelectingText: true }),
    ).toBe("ignore");
    expect(
      decide({ region: "belowLastBlock", isSelectingText: true }),
    ).toBe("ignore");
  });

  it("ignores header chrome, in-block, and other regions", () => {
    expect(decide({ region: "headerChrome" })).toBe("ignore");
    expect(decide({ region: "onBlock" })).toBe("ignore");
    expect(decide({ region: "other" })).toBe("ignore");
  });

  it("ignores when the document has no blocks", () => {
    expect(decide({ hasBlocks: false })).toBe("ignore");
  });

  it("stays idempotent once a trailing empty paragraph exists", () => {
    const first = decide({ lastBlockIsEmptyParagraph: false });
    const second = decide({ lastBlockIsEmptyParagraph: true });
    const third = decide({ lastBlockIsEmptyParagraph: true });

    expect(first).toBe("appendAndFocus");
    expect(second).toBe("focusLast");
    expect(third).toBe("focusLast");
  });
});
