import { describe, expect, it } from "vitest";
import {
  formatLineRangeLabel,
  normalizeRepoPath,
  sliceFileLines,
  splitRepoPath,
} from "./repoFileRef";

describe("repoFileRef", () => {
  it("normalizes paths and splits file/dir", () => {
    expect(normalizeRepoPath("src\\lib\\queue.ts")).toBe("src/lib/queue.ts");
    expect(splitRepoPath("src/lib/queue.ts")).toEqual({
      fileName: "queue.ts",
      dirPath: "src/lib/",
    });
  });

  it("formats line ranges", () => {
    expect(formatLineRangeLabel(42, 58)).toBe("lines 42–58");
    expect(formatLineRangeLabel(10, 10)).toBe("line 10");
    expect(formatLineRangeLabel(undefined)).toBeNull();
  });

  it("slices file content by 1-based lines", () => {
    const content = "a\nb\nc\nd\ne";
    expect(sliceFileLines(content, 2, 4)).toEqual({
      start: 2,
      lines: ["b", "c", "d"],
    });
  });
});
