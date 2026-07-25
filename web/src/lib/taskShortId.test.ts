import { describe, expect, it } from "vitest";
import { shortId, taskDisplayRef } from "./taskShortId";

describe("shortId", () => {
  it("returns the first 8 characters for longer ids", () => {
    expect(shortId("abcdef12-3456-7890")).toBe("abcdef12");
  });

  it("returns the full id when length is at most 8", () => {
    expect(shortId("abc")).toBe("abc");
    expect(shortId("12345678")).toBe("12345678");
  });
});

describe("taskDisplayRef", () => {
  it("prefers the per-project #N ref when number is set", () => {
    expect(taskDisplayRef({ id: "abcdef12-3456-7890", number: 42 })).toBe(
      "#42",
    );
  });

  it("falls back to the shortened UUID when number is missing", () => {
    expect(taskDisplayRef({ id: "abcdef12-3456-7890" })).toBe("abcdef12");
  });

  it("falls back to the shortened UUID when number is null", () => {
    expect(
      taskDisplayRef({ id: "abcdef12-3456-7890", number: null }),
    ).toBe("abcdef12");
  });

  it("ignores non-finite numeric values defensively", () => {
    expect(
      taskDisplayRef({ id: "abcdef12-3456-7890", number: Number.NaN }),
    ).toBe("abcdef12");
  });
});
