import { describe, expect, it } from "vitest";
import { wordCountFromHtml } from "./promptEditorPageMeta";

describe("wordCountFromHtml", () => {
  it("returns 0 for empty or whitespace-only HTML", () => {
    expect(wordCountFromHtml("")).toBe(0);
    expect(wordCountFromHtml("<p></p>")).toBe(0);
    expect(wordCountFromHtml("<p>   </p>")).toBe(0);
  });

  it("counts words from BlockNote-style HTML prose", () => {
    expect(wordCountFromHtml("<p>Hello world</p>")).toBe(2);
    expect(
      wordCountFromHtml("<p>One</p><p>two <strong>three</strong></p>"),
    ).toBe(3);
  });
});
