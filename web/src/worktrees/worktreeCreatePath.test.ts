import { describe, expect, it } from "vitest";
import { joinWorktreeCreatePath } from "./worktreeCreatePath";

describe("joinWorktreeCreatePath", () => {
  it("joins parent and folder on unix paths", () => {
    expect(joinWorktreeCreatePath("/repo", "feature-a")).toBe("/repo/feature-a");
  });

  it("joins parent and folder on windows paths", () => {
    expect(joinWorktreeCreatePath("C:\\dev\\repo", "feature-a")).toBe("C:\\dev\\repo\\feature-a");
  });

  it("trims trailing slashes from parent and leading slashes from folder", () => {
    expect(joinWorktreeCreatePath("/repo/", "/feature-a/")).toBe("/repo/feature-a");
  });

  it("returns null when parent or folder is empty", () => {
    expect(joinWorktreeCreatePath("", "feature")).toBeNull();
    expect(joinWorktreeCreatePath("/repo", "")).toBeNull();
    expect(joinWorktreeCreatePath("  ", "feature")).toBeNull();
  });

  it("rejects folder names with path separators", () => {
    expect(joinWorktreeCreatePath("/repo", "a/b")).toBeNull();
    expect(joinWorktreeCreatePath("/repo", "a\\b")).toBeNull();
  });
});
