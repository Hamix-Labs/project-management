import { describe, expect, it } from "vitest";
import { stripEphemeralEditorAttrs } from "./stripEphemeralEditorAttrs";

describe("stripEphemeralEditorAttrs", () => {
  it("removes UniqueID data-id attributes from TipTap HTML", () => {
    expect(stripEphemeralEditorAttrs('<p data-id="abc123"></p>')).toBe("<p></p>");
    expect(
      stripEphemeralEditorAttrs('<h2 data-id="x">Hello</h2>'),
    ).toBe("<h2>Hello</h2>");
  });
});
