import { describe, expect, it } from "vitest";
import {
  MAX_SELECTED_PROJECT_CONTEXT_ITEMS,
  mergeProjectContextSelection,
  projectContextShortId,
} from "./projectContextRefs";

describe("projectContextShortId", () => {
  it("strips dashes and lowercases UUIDs to a 6-char prefix", () => {
    expect(projectContextShortId("A1B2C3-D4E5F6")).toBe("a1b2c3");
  });

  it("returns empty for empty input", () => {
    expect(projectContextShortId("")).toBe("");
    expect(projectContextShortId("   ")).toBe("");
  });

  it("falls back to the trimmed id when only punctuation remains", () => {
    expect(projectContextShortId("---")).toBe("---");
  });

  it("preserves alphanumerics from arbitrary ids", () => {
    expect(projectContextShortId("ctx-risk")).toBe("ctxris");
  });
});

describe("mergeProjectContextSelection", () => {
  it("appends only new ids and keeps order", () => {
    expect(mergeProjectContextSelection(["a", "b"], ["b", "c"])).toEqual([
      "a",
      "b",
      "c",
    ]);
  });

  it("returns the existing list unchanged when nothing new is added", () => {
    const existing = ["a", "b"];
    const out = mergeProjectContextSelection(existing, ["a"]);
    expect(out).toEqual(existing);
    expect(out).not.toBe(existing);
  });

  it("respects the server-side selection cap", () => {
    const seed: string[] = [];
    for (let i = 0; i < MAX_SELECTED_PROJECT_CONTEXT_ITEMS; i += 1) {
      seed.push(`existing-${i}`);
    }
    const merged = mergeProjectContextSelection(seed, ["new-1", "new-2"]);
    expect(merged.length).toBe(MAX_SELECTED_PROJECT_CONTEXT_ITEMS);
    expect(merged).toEqual(seed);
  });

  it("trims incoming ids and skips empty strings", () => {
    expect(mergeProjectContextSelection([], [" a ", "", "  "])).toEqual(["a"]);
  });
});
