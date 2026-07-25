import { describe, expect, it } from "vitest";
import { shortId } from "./taskShortId";

describe("shortId", () => {
  it("returns the first 8 characters for longer ids", () => {
    expect(shortId("abcdef12-3456-7890")).toBe("abcdef12");
  });

  it("returns the full id when length is at most 8", () => {
    expect(shortId("abc")).toBe("abc");
    expect(shortId("12345678")).toBe("12345678");
  });
});
