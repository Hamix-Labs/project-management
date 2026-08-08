import { describe, expect, it } from "vitest";
import { deriveWordCountLabel } from "./promptEditorPageViewModel";

describe("deriveWordCountLabel", () => {
  it("returns an em dash while not ready", () => {
    expect(deriveWordCountLabel(false, "<p>hello world</p>")).toBe("—");
  });

  it("returns 0 words for empty ready html", () => {
    expect(deriveWordCountLabel(true, "")).toBe("0 words");
    expect(deriveWordCountLabel(true, "<p></p>")).toBe("0 words");
  });

  it("returns an approximate count for BlockNote-style html", () => {
    expect(deriveWordCountLabel(true, "<p>Hello world</p>")).toBe("~2 words");
    expect(
      deriveWordCountLabel(
        true,
        "<p>One</p><p>two <strong>three</strong></p>",
      ),
    ).toBe("~3 words");
  });
});
