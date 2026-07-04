import { describe, expect, it } from "vitest";
import {
  repositoryDisplayName,
  repositoryMatchesSearchQuery,
  repositoryPathsEquivalent,
  shouldShowWorktreePath,
  splitWorktreePath,
  worktreeMatchesSearchQuery,
  worktreePathLabel,
} from "./repositoryDisplay";

describe("repositoryDisplayName", () => {
  it("returns the last path segment", () => {
    expect(repositoryDisplayName("C:/Users/dev/Documents/hamix")).toBe("hamix");
    expect(repositoryDisplayName("/repo/main")).toBe("main");
  });
});

describe("repositoryPathsEquivalent", () => {
  it("treats equivalent paths as equal regardless of separators", () => {
    expect(
      repositoryPathsEquivalent(
        "C:/Users/dev/OneDrive/Documents/hamix",
        "C:\\Users\\dev\\OneDrive\\Documents\\hamix",
      ),
    ).toBe(true);
  });
});

describe("splitWorktreePath", () => {
  it("splits parent directory from the final segment", () => {
    expect(splitWorktreePath("C:\\Users\\dev\\Documents\\hamix")).toEqual({
      parent: "C:\\Users\\dev\\Documents\\",
      base: "hamix",
    });
    expect(splitWorktreePath("/repo/feature")).toEqual({
      parent: "/repo/",
      base: "feature",
    });
  });
});

describe("shouldShowWorktreePath", () => {
  it("hides path when it matches the repository header path", () => {
    expect(shouldShowWorktreePath("/repo/main", "/repo/main")).toBe(false);
    expect(shouldShowWorktreePath("/repo/feature", "/repo/main")).toBe(true);
  });
});

describe("worktreePathLabel", () => {
  it("shows sibling folder name when worktrees share a parent directory", () => {
    expect(
      worktreePathLabel(
        "C:/Users/dev/Documents/Hamix-wt-03",
        "C:/Users/dev/Documents/Hamix-wt-polish",
      ),
    ).toBe("Hamix-wt-03");
  });

  it("shows a relative suffix when the worktree is under the repository path", () => {
    expect(worktreePathLabel("/repo/main/feature", "/repo/main")).toBe("feature");
  });
});

describe("repositoryMatchesSearchQuery", () => {
  const repo = {
    path: "/repo/hamix",
    host_path: "C:/Users/dev/Documents/hamix",
  };

  it("matches display name, path, and host path", () => {
    expect(repositoryMatchesSearchQuery(repo, "")).toBe(true);
    expect(repositoryMatchesSearchQuery(repo, "hamix")).toBe(true);
    expect(repositoryMatchesSearchQuery(repo, "/repo/hamix")).toBe(true);
    expect(repositoryMatchesSearchQuery(repo, "documents")).toBe(true);
    expect(repositoryMatchesSearchQuery(repo, "nomatch")).toBe(false);
  });
});

describe("worktreeMatchesSearchQuery", () => {
  const worktree = { name: "Hamix-wt-03", path: "/repo/Hamix-wt-03" };

  it("matches name, path, and branch", () => {
    expect(worktreeMatchesSearchQuery(worktree, "hamix-test-03", "")).toBe(true);
    expect(worktreeMatchesSearchQuery(worktree, "hamix-test-03", "wt-03")).toBe(true);
    expect(worktreeMatchesSearchQuery(worktree, "hamix-test-03", "hamix-test")).toBe(true);
    expect(worktreeMatchesSearchQuery(worktree, "hamix-test-03", "/repo/hamix")).toBe(true);
    expect(worktreeMatchesSearchQuery(worktree, undefined, "nomatch")).toBe(false);
  });
});
