import { describe, expect, it } from "vitest";
import { htmlToInitialBlocks } from "./promptEditorHtml";

describe("htmlToInitialBlocks", () => {
  it("parses paragraph HTML into blocks with visible text", () => {
    const { blocks, usedFallback } = htmlToInitialBlocks(
      "<p>Hello world from the prompt</p>",
    );
    expect(usedFallback).toBe(false);
    expect(blocks.length).toBeGreaterThan(0);
    const json = JSON.stringify(blocks);
    expect(json).toContain("Hello world from the prompt");
  });

  it("migrates plain text into paragraph blocks", () => {
    const { blocks, usedFallback } = htmlToInitialBlocks(
      "Line one\n\nLine two has words",
    );
    expect(usedFallback).toBe(false);
    expect(JSON.stringify(blocks)).toMatch(/Line one/);
  });

  it("never returns empty blocks for non-empty input", () => {
    const { blocks, usedFallback } = htmlToInitialBlocks(
      "<div>Orphan rich text without blocknote tags</div><p>Still here</p>",
    );
    expect(blocks.length).toBeGreaterThan(0);
    expect(JSON.stringify(blocks).length).toBeGreaterThan(10);
    void usedFallback;
  });

  it("returns a single empty paragraph for empty input", () => {
    const { blocks, usedFallback } = htmlToInitialBlocks("");
    expect(usedFallback).toBe(false);
    expect(blocks.length).toBeGreaterThan(0);
  });

  it("round-trips multi-paragraph content with non-empty text", () => {
    const html =
      "<p>One alpha</p><p>Two <strong>beta</strong> gamma</p><pre><code>const x = 1;</code></pre>";
    const { blocks } = htmlToInitialBlocks(html);
    const text = JSON.stringify(blocks);
    expect(text).toMatch(/One alpha/);
    expect(text).toMatch(/beta|gamma|Two/);
  });
});
